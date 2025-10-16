package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/jobs"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetWalletBalance retrieves the wallet balances for a user (both NGN and GHS)
func GetWalletBalance(userID uint) ([]types.WalletBalance, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Find or create wallets
	wallets, err := findOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	var balances []types.WalletBalance
	for _, wallet := range wallets {
		balance := types.WalletBalance{
			ID:                wallet.ID,
			UserID:            wallet.UserID,
			Balance:           utils.RoundCurrency(wallet.Balance),
			Currency:          wallet.Currency,
			WalletID:          wallet.WalletID,
			TotalDeposits:     utils.RoundCurrency(wallet.TotalDeposits),
			TotalWithdrawals:  utils.RoundCurrency(wallet.TotalWithdrawals),
			TotalConversions:  utils.RoundCurrency(wallet.TotalConversions),
			IsActive:          wallet.IsActive,
			LastTransactionAt: wallet.LastTransactionAt,
			UpdatedAt:         wallet.UpdatedAt,
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// TopUpWallet initiates a wallet top-up
func TopUpWallet(userID uint, req types.TopUpRequest) (*types.TopUpResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Validate request
	if err := validateTopUpRequest(req); err != nil {
		return nil, err
	}

	// Ensure wallets exist
	_, err := findOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to access wallets: %w", err)
	}

	// Get current exchange rate
	currentExchangeRate, err := GetExchangeRates()
	if err != nil {
		return nil, fmt.Errorf("exchange rate not found: %w", err)
	}

	// Create transaction record
	now := time.Now()
	transactionIdx, err := libs.SecureRandomNumber(12)
	if err != nil {
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}
	transactionID := fmt.Sprintf("TOP%d", transactionIdx)
	reference := req.PaymentReference
	if reference == "" {
		reference = generateTransactionReference("TOPUP")
	}

	// Determine payment type based on method
	var paymentType models.PaymentType
	if req.PaymentMethod == "bank" {
		paymentType = models.PaymentTypeBank
	} else {
		paymentType = models.PaymentTypeMomo
	}

	formattedCurrency := utils.FormatCurrency(req.Amount, req.Currency)

	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		PaymentType:     paymentType,
		Status:          models.TransactionPending,
		TransactionType: models.Deposit,
		Reference:       reference,
		Direction:       getDepositDirection(req.Currency),
		CurrentRate:     currentExchangeRate.Rates[string(getDepositDirection(req.Currency))],
		Description:     fmt.Sprintf("Wallet top-up of %s %s using %s", formattedCurrency, req.Currency, req.PaymentMethod),
		TransactionDetails: models.TransactionDetails{
			FromCurrency: req.Currency,
			ToCurrency:   req.Currency,
			FromAmount:   req.Amount,
			ToAmount:     req.Amount,
			MethodOfPayment: func() string {
				if req.IsDirectPayment {
					return "checkout"
				}
				return "wallet"
			}(),
		},
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Send email notification to admins for new wallet top-up transaction
	adminEmails, err := GetAdminEmails()
	if err != nil {
		// Log the error but don't fail the transaction
		fmt.Printf("Warning: Failed to get admin emails for wallet top-up notification: %v\n", err)
	} else if len(adminEmails) > 0 {
		// Get full user details for email
		var fullUser models.User
		if err := database.DB.First(&fullUser, userID).Error; err == nil {
			emailClient := jobs.NewEmailJobClient()
			defer emailClient.Close()
			emailClient.EnqueueNewDepositAdmin(adminEmails, fullUser.FirstName+" "+fullUser.LastName, fullUser.Email, transaction)
		}
	}

	notificationClient := jobs.NewNotificationJobClient()
	title := "Wallet Top Up"
	message := fmt.Sprintf("Your wallet top-up of %s %s is being processed", utils.FormatCurrency(req.Amount, req.Currency), req.Currency)
	notificationClient.EnqueueCreateNotification(
		userID,
		models.NotificationType("topup"),
		title,
		message,
	)

	return &types.TopUpResponse{
		TransactionID:    transactionID,
		Amount:           utils.RoundCurrency(req.Amount),
		Currency:         req.Currency,
		PaymentMethod:    req.PaymentMethod,
		Status:           "pending",
		PaymentReference: reference,
		CreatedAt:        now,
	}, nil
}

// CalculateWithdrawalFee calculates withdrawal fee with 1.5% rate and 20 GHS cap
func CalculateWithdrawalFee(amount float64, currency string) types.WithdrawalFeeCalculation {
	const feeRate = 0.015  // 1.5%
	const feeCapGHS = 20.0 // 20 GHS cap

	calculatedFee := amount * feeRate
	finalFee := calculatedFee

	// Apply fee cap for GHS currency
	if currency == "GHS" && calculatedFee > feeCapGHS {
		finalFee = feeCapGHS
	}

	// For NGN, convert the GHS cap to NGN equivalent if needed
	if currency == "NGN" {
		// Get current exchange rate to convert GHS cap to NGN
		exchangeRates, err := GetExchangeRates()
		if err == nil && exchangeRates.Rates["GHS-NGN"] > 0 {
			feeCapNGN := feeCapGHS * exchangeRates.Rates["GHS-NGN"]
			if calculatedFee > feeCapNGN {
				finalFee = feeCapNGN
			}
		}
	}

	netAmount := amount - finalFee

	return types.WithdrawalFeeCalculation{
		Amount:        utils.RoundCurrency(amount),
		Currency:      currency,
		FeeRate:       feeRate,
		CalculatedFee: utils.RoundCurrency(calculatedFee),
		FinalFee:      utils.RoundCurrency(finalFee),
		FeeCap:        feeCapGHS,
		NetAmount:     utils.RoundCurrency(netAmount),
	}
}

// ValidateWithdrawalLimits validates withdrawal limits and balance
func ValidateWithdrawalLimits(userID uint, amount float64, currency string) error {
	if amount <= 0 {
		return errors.New("withdrawal amount must be greater than zero")
	}

	// Get user's wallet balance
	wallets, err := findOrCreateWallet(userID)
	if err != nil {
		return fmt.Errorf("failed to access wallets: %w", err)
	}

	// Find the wallet for the specific currency
	var targetWallet *models.Wallet
	for _, wallet := range wallets {
		if wallet.Currency == currency {
			targetWallet = &wallet
			break
		}
	}

	if targetWallet == nil {
		return fmt.Errorf("wallet not found for currency %s", currency)
	}

	// Calculate fee and check if user has enough balance including fee
	feeCalc := CalculateWithdrawalFee(amount, currency)
	totalRequired := amount + feeCalc.FinalFee

	if targetWallet.Balance < totalRequired {
		return fmt.Errorf("insufficient balance. Required: %s %.2f (including fee: %s %.2f), Available: %s %.2f",
			currency, totalRequired, currency, feeCalc.FinalFee, currency, targetWallet.Balance)
	}

	return nil
}

// WithdrawFromWallet initiates a wallet withdrawal with fee calculation
func WithdrawFromWallet(userID uint, req types.WithdrawRequest) (*types.WithdrawResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Validate request
	if err := validateWithdrawRequest(req); err != nil {
		return nil, err
	}

	// Validate withdrawal limits and balance
	if err := ValidateWithdrawalLimits(userID, req.Amount, req.Currency); err != nil {
		return nil, err
	}

	// Calculate withdrawal fee
	feeCalc := CalculateWithdrawalFee(req.Amount, req.Currency)
	totalDeduction := req.Amount + feeCalc.FinalFee

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Deduct total amount (withdrawal + fee) from wallet
	err := updateWalletBalance(tx, userID, req.Currency, -totalDeduction, "withdrawal")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Get current exchange rate
	currentExchangeRate, err := GetExchangeRates()
	if err != nil {
		tx.Rollback()
		return nil, errors.New("exchange rate not found")
	}

	// Create transaction record
	now := time.Now()
	transactionIdx, err := libs.SecureRandomNumber(12)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}
	transactionID := fmt.Sprintf("WTH%d", transactionIdx)
	reference := generateTransactionReference("WITHDRAW")

	// Determine payment type based on method
	var paymentType models.PaymentType
	if req.WithdrawalMethod == "bank" {
		paymentType = models.PaymentTypeBank
	} else {
		paymentType = models.PaymentTypeMomo
	}

	// Create fee calculation details for storage
	feeDetails := fmt.Sprintf("Amount: %s %.2f, Fee Rate: %.1f%%, Calculated Fee: %s %.2f, Final Fee: %s %.2f, Net Amount: %s %.2f",
		req.Currency, feeCalc.Amount, feeCalc.FeeRate*100, req.Currency, feeCalc.CalculatedFee,
		req.Currency, feeCalc.FinalFee, req.Currency, feeCalc.NetAmount)

	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		PaymentType:     paymentType,
		Status:          models.TransactionPending,
		TransactionType: models.Withdrawal,
		Reference:       reference,
		Direction:       getWithdrawalDirection(req.Currency),
		CurrentRate:     currentExchangeRate.Rates[string(getWithdrawalDirection(req.Currency))],
		WithdrawalFee:   feeCalc.FinalFee,
		FeeCalculation:  feeDetails,
		Description:     fmt.Sprintf("Withdraw %s %.2f via %s (Fee: %s %.2f)", req.Currency, req.Amount, req.WithdrawalMethod, req.Currency, feeCalc.FinalFee),
		TransactionDetails: models.TransactionDetails{
			FromCurrency:    req.Currency,
			ToCurrency:      req.Currency,
			FromAmount:      req.Amount,
			ToAmount:        feeCalc.NetAmount, // Net amount after fee
			MethodOfPayment: req.WithdrawalMethod,
			// Store account details in appropriate fields
			AccountNumber: func() string {
				if req.AccountDetails.AccountNumber != "" {
					return req.AccountDetails.AccountNumber
				}
				return req.AccountDetails.PhoneNumber
			}(),
			BankName:      req.AccountDetails.BankName,
			Network:       req.AccountDetails.Network,
			RecipientName: req.AccountDetails.AccountName,
		},
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Send notification
	notificationClient := jobs.NewNotificationJobClient()
	title := "Withdrawal Request"
	message := fmt.Sprintf("Your withdrawal request of %s %.2f is being processed (Fee: %s %.2f)",
		req.Currency, req.Amount, req.Currency, feeCalc.FinalFee)
	notificationClient.EnqueueCreateNotification(
		userID,
		models.NotificationType("withdrawal"),
		title,
		message,
	)

	return &types.WithdrawResponse{
		TransactionID:    transactionID,
		Amount:           utils.RoundCurrency(req.Amount),
		Currency:         req.Currency,
		WithdrawalMethod: req.WithdrawalMethod,
		Status:           models.TransactionPending,
		AccountDetails:   req.AccountDetails,
		Fee:              feeCalc.FinalFee,
		NetAmount:        feeCalc.NetAmount,
		CreatedAt:        now,
	}, nil
}

// GetWalletHistory retrieves wallet transaction history
func GetWalletHistory(userID uint32, pagination types.PaginationRequest, txType, status string) ([]types.WalletHistoryItem, *types.PaginationResponse, error) {
	if userID == 0 {
		return nil, nil, errors.New("user ID is required")
	}

	// Build query
	query := database.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)

	if txType != "" && isValidTransactionType(txType) {
		if txType == "deposit" || txType == "withdrawal" {
			query = query.Where("transaction_type = ?", txType)
		}
	}

	if status != "" && isValidTransactionStatus(status) {
		query = query.Where("status = ?", status)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Calculate pagination
	page := pagination.Page
	if page <= 0 {
		page = 1
	}
	limit := pagination.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Find transactions
	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find transactions: %w", err)
	}

	// Convert to history items
	var historyItems []types.WalletHistoryItem
	for _, tx := range transactions {
		historyItems = append(historyItems, types.WalletHistoryItem{
			TransactionID: tx.TransactionID,
			Type:          string(tx.TransactionType),
			Status:        string(tx.Status),
			Direction:     string(tx.Direction),
			Description:   tx.Description,
			Reference:     tx.Reference,
			CreatedAt:     tx.CreatedAt,
		})
	}

	// Create pagination response
	paginationResp := &types.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	return historyItems, paginationResp, nil
}

// UpdateWalletAfterPayment updates wallet balance after successful payment
func UpdateWalletAfterPayment(userID uint, currency string, amount float64) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	return updateWalletBalance(database.DB, userID, currency, amount, "deposit")
}

// Helper functions

// findOrCreateWallet finds existing wallets or creates new ones (NGN, GHS, USD, XOF)
func findOrCreateWallet(userID uint) ([]models.Wallet, error) {
	var wallets []models.Wallet

	err := database.DB.Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find wallets: %w", err)
	}

	// Check which wallets exist
	supportedCurrencies := []string{"NGN", "GHS", "USD", "XOF"}
	existingCurrencies := make(map[string]bool)

	for _, wallet := range wallets {
		existingCurrencies[wallet.Currency] = true
	}

	// Create missing wallets for all supported currencies
	for _, currency := range supportedCurrencies {
		if !existingCurrencies[currency] {
			var walletIDLength int
			switch currency {
			case "NGN":
				walletIDLength = 12
			case "GHS":
				walletIDLength = 10
			case "USD":
				walletIDLength = 11
			case "XOF":
				walletIDLength = 13
			default:
				walletIDLength = 12
			}

			newWallet := models.Wallet{
				UserID:           userID,
				Currency:         currency,
				Balance:          0.0,
				WalletID:         libs.GenerateRandomLengthNumbers(walletIDLength),
				TotalDeposits:    0.0,
				TotalWithdrawals: 0.0,
				TotalConversions: 0.0,
				IsActive:         true,
			}

			if err := database.DB.Create(&newWallet).Error; err != nil {
				return nil, fmt.Errorf("failed to create %s wallet: %w", currency, err)
			}
			wallets = append(wallets, newWallet)
		}
	}

	return wallets, nil
}

// updateWalletBalance updates wallet balance atomically for specific currency
func updateWalletBalance(tx *gorm.DB, userID uint, currency string, amount float64, txType string) error {
	var wallet models.Wallet

	// Find wallet with lock for specific currency
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ? AND currency = ?", userID, currency).First(&wallet).Error; err != nil {
		return fmt.Errorf("failed to find %s wallet: %w", currency, err)
	}

	// Update balance
	roundedAmount := utils.RoundCurrency(amount)
	wallet.Balance += roundedAmount

	// Update totals based on transaction type
	if txType == "deposit" && amount > 0 {
		wallet.TotalDeposits += roundedAmount
	} else if txType == "withdrawal" && amount < 0 {
		wallet.TotalWithdrawals += -roundedAmount
	} else if txType == "conversion" {
		if amount < 0 {
			wallet.TotalConversions += -roundedAmount
		}
	}

	// Update timestamps
	now := time.Now()
	wallet.UpdatedAt = now
	wallet.LastTransactionAt = &now

	// Save wallet
	if err := tx.Save(&wallet).Error; err != nil {
		return fmt.Errorf("failed to update %s wallet: %w", currency, err)
	}

	return nil
}

// getBalanceByCurrency gets balance for specific currency from wallet slice
func getBalanceByCurrency(wallets []models.Wallet, currency string) float64 {
	for _, wallet := range wallets {
		if wallet.Currency == currency {
			return wallet.Balance
		}
	}
	return 0
}

// getDepositDirection gets transaction direction for deposit
func getDepositDirection(currency string) models.TransactionDirection {
	switch currency {
	case "NGN":
		return "DEPOSIT-NGN"
	case "GHS":
		return "DEPOSIT-GHS"
	default:
		return ""
	}
}

// getWithdrawalDirection gets transaction direction for withdrawal
func getWithdrawalDirection(currency string) models.TransactionDirection {
	switch currency {
	case "NGN":
		return "WITHDRAWAL-NGN"
	case "GHS":
		return "WITHDRAWAL-GHS"
	default:
		return ""
	}
}

// generateTransactionReference generates a unique transaction reference
func generateTransactionReference(prefix string) string {
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), uuid.New().String()[:8])
}

// validateTopUpRequest validates top-up request
func validateTopUpRequest(req types.TopUpRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if req.Currency != "NGN" && req.Currency != "GHS" {
		return errors.New("invalid currency. Must be NGN or GHS")
	}

	if req.PaymentMethod != "bank" && req.PaymentMethod != "momo" {
		return errors.New("invalid payment method. Must be bank or momo")
	}

	return nil
}

// validateWithdrawRequest validates withdrawal request
func validateWithdrawRequest(req types.WithdrawRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if req.Currency != "NGN" && req.Currency != "GHS" {
		return errors.New("invalid currency. Must be NGN or GHS")
	}

	fmt.Println(req.WithdrawalMethod)

	if req.WithdrawalMethod != "bank" && req.WithdrawalMethod != "momo" {
		return errors.New("invalid withdrawal method. Must be bank or momo")
	}

	// Validate account details based on withdrawal method
	if req.WithdrawalMethod == "bank" {
		if req.AccountDetails.AccountNumber == "" {
			return errors.New("account number is required for bank withdrawal")
		}
		if req.AccountDetails.AccountName == "" {
			return errors.New("account name is required for bank withdrawal")
		}
		if req.AccountDetails.BankName == "" {
			return errors.New("bank name is required for bank withdrawal")
		}
	} else if req.WithdrawalMethod == "momo" {
		if req.AccountDetails.PhoneNumber == "" {
			return errors.New("phone number is required for mobile money withdrawal")
		}
		if req.AccountDetails.Network == "" {
			return errors.New("network is required for mobile money withdrawal")
		}
	}

	return nil
}

// isValidTransactionType checks if transaction type is valid
func isValidTransactionType(txType string) bool {
	validTypes := []string{"deposit", "withdrawal", "conversion", "transfer"}
	for _, t := range validTypes {
		if t == txType {
			return true
		}
	}
	return false
}

// isValidTransactionStatus checks if transaction status is valid
func isValidTransactionStatus(status string) bool {
	validStatuses := []string{"pending", "completed", "failed"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetTopUpDetails retrieves topup transaction details
func GetTopUpDetails(userID uint, transactionID string) (*types.TopUpResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	if transactionID == "" {
		return nil, errors.New("transaction ID is required")
	}

	var transaction models.Transaction
	err := database.DB.Where("user_id = ? AND transaction_id = ? AND transaction_type = ?",
		userID, transactionID, models.Deposit).
		Preload("TransactionDetails").
		First(&transaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("topup transaction not found")
		}
		return nil, fmt.Errorf("failed to fetch topup details: %w", err)
	}

	// Convert to response format
	response := &types.TopUpResponse{
		TransactionID:    transaction.TransactionID,
		Amount:           utils.RoundCurrency(transaction.TransactionDetails.FromAmount),
		Currency:         transaction.TransactionDetails.FromCurrency,
		PaymentMethod:    string(transaction.PaymentType),
		Status:           string(transaction.Status),
		PaymentReference: transaction.Reference,
		CreatedAt:        transaction.CreatedAt,
	}

	return response, nil
}

// GetWithdrawalDetails retrieves withdrawal transaction details
func GetWithdrawalDetails(userID uint, transactionID string) (*types.WithdrawResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	if transactionID == "" {
		return nil, errors.New("transaction ID is required")
	}

	var transaction models.Transaction
	err := database.DB.Where("user_id = ? AND transaction_id = ? AND transaction_type = ?",
		userID, transactionID, models.Withdrawal).
		Preload("TransactionDetails").
		First(&transaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("withdrawal transaction not found")
		}
		return nil, fmt.Errorf("failed to fetch withdrawal details: %w", err)
	}

	// Convert to response format
	response := &types.WithdrawResponse{
		TransactionID:    transaction.TransactionID,
		Amount:           utils.RoundCurrency(transaction.TransactionDetails.FromAmount),
		Currency:         transaction.TransactionDetails.FromCurrency,
		WithdrawalMethod: string(transaction.PaymentType),
		Status:           transaction.Status,
		AccountDetails: types.WithdrawalAccountDetails{
			AccountNumber: transaction.TransactionDetails.AccountNumber,
			AccountName:   transaction.TransactionDetails.RecipientName,
			BankName:      transaction.TransactionDetails.BankName,
			PhoneNumber:   transaction.TransactionDetails.PhoneNumber,
			Network:       transaction.TransactionDetails.Network,
		},
		Fee:       utils.RoundCurrency(transaction.WithdrawalFee),
		NetAmount: utils.RoundCurrency(transaction.TransactionDetails.ToAmount),
		CreatedAt: transaction.CreatedAt,
	}

	return response, nil
}

// GetUserWithdrawals retrieves user's withdrawal history
func GetUserWithdrawals(userID uint, req types.GetWithdrawalsRequest) (*types.GetWithdrawalsResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Build query
	query := database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ?", userID, models.Withdrawal).
		Preload("TransactionDetails")

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Currency != "" {
		query = query.Joins("JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
			Where("transaction_details.from_currency = ?", req.Currency)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count withdrawals: %w", err)
	}

	// Calculate pagination
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Find transactions
	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to find withdrawals: %w", err)
	}

	// Convert to response format
	var withdrawals []types.WithdrawResponse
	for _, tx := range transactions {
		withdrawal := types.WithdrawResponse{
			TransactionID:    tx.TransactionID,
			Amount:           utils.RoundCurrency(tx.TransactionDetails.FromAmount),
			Currency:         tx.TransactionDetails.FromCurrency,
			WithdrawalMethod: string(tx.PaymentType),
			Status:           tx.Status,
			AccountDetails: types.WithdrawalAccountDetails{
				AccountNumber: tx.TransactionDetails.AccountNumber,
				AccountName:   tx.TransactionDetails.RecipientName,
				BankName:      tx.TransactionDetails.BankName,
				PhoneNumber:   tx.TransactionDetails.PhoneNumber,
				Network:       tx.TransactionDetails.Network,
			},
			Fee:       utils.RoundCurrency(tx.WithdrawalFee),
			NetAmount: utils.RoundCurrency(tx.TransactionDetails.ToAmount),
			CreatedAt: tx.CreatedAt,
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	// Create pagination response
	totalPages := (total + int64(limit) - 1) / int64(limit)

	return &types.GetWithdrawalsResponse{
		Withdrawals: withdrawals,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  int(totalPages),
	}, nil
}
