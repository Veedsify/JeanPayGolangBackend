package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetWalletBalance retrieves the wallet balances for a user (both NGN and GHS)
func GetWalletBalance(userID uint) ([]types.WalletBalance, error) {
	fmt.Println(userID)
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

// WithdrawFromWallet initiates a wallet withdrawal
func WithdrawFromWallet(userID uint, req types.WithdrawRequest) (*types.WithdrawResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Validate request
	if err := validateWithdrawRequest(req); err != nil {
		return nil, err
	}

	// Get wallets and check balance
	wallets, err := findOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to access wallets: %w", err)
	}

	// Find the wallet for the specific currency and check balance
	var targetWallet *models.Wallet
	for _, wallet := range wallets {
		if wallet.Currency == req.Currency {
			targetWallet = &wallet
			break
		}
	}

	if targetWallet == nil {
		return nil, fmt.Errorf("wallet not found for currency %s", req.Currency)
	}

	if targetWallet.Balance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Deduct amount from wallet
	err = updateWalletBalance(tx, userID, req.Currency, -req.Amount, "withdrawal")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Create transaction record
	now := time.Now()
	transactionID := uuid.New().String()
	reference := generateTransactionReference("WITHDRAW")

	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		Status:          "pending",
		TransactionType: "withdrawal",
		Reference:       reference,
		Direction:       getWithdrawalDirection(req.Currency),
		Description:     fmt.Sprintf("Withdraw %s %.2f via %s", req.Currency, req.Amount, req.WithdrawalMethod),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &types.WithdrawResponse{
		TransactionID:    transactionID,
		Amount:           utils.RoundCurrency(req.Amount),
		Currency:         req.Currency,
		WithdrawalMethod: req.WithdrawalMethod,
		Status:           "pending",
		AccountDetails:   req.AccountDetails,
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

// findOrCreateWallet finds existing wallets or creates new ones (NGN and GHS)
func findOrCreateWallet(userID uint) ([]models.Wallet, error) {
	var wallets []models.Wallet

	err := database.DB.Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find wallets: %w", err)
	}

	// Check if we have both NGN and GHS wallets
	hasNGN := false
	hasGHS := false
	for _, wallet := range wallets {
		if wallet.Currency == "NGN" {
			hasNGN = true
		}
		if wallet.Currency == "GHS" {
			hasGHS = true
		}
	}

	// Create missing wallets
	if !hasNGN {
		ngnWallet := models.Wallet{
			UserID:           userID,
			Currency:         "NGN",
			Balance:          0.0,
			WalletID:         libs.GenerateRandomLengthNumbers(12), // NGN wallet ID
			TotalDeposits:    0.0,
			TotalWithdrawals: 0.0,
			TotalConversions: 0.0,
			IsActive:         true,
		}
		if err := database.DB.Create(&ngnWallet).Error; err != nil {
			return nil, fmt.Errorf("failed to create NGN wallet: %w", err)
		}
		wallets = append(wallets, ngnWallet)
	}

	if !hasGHS {
		ghsWallet := models.Wallet{
			UserID:           userID,
			Currency:         "GHS",
			Balance:          0.0,
			WalletID:         libs.GenerateRandomLengthNumbers(10), // GHS wallet ID
			TotalDeposits:    0.0,
			TotalWithdrawals: 0.0,
			TotalConversions: 0.0,
			IsActive:         true,
		}
		if err := database.DB.Create(&ghsWallet).Error; err != nil {
			return nil, fmt.Errorf("failed to create GHS wallet: %w", err)
		}
		wallets = append(wallets, ghsWallet)
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

	if req.WithdrawalMethod != "bank" && req.WithdrawalMethod != "momo" {
		return errors.New("invalid withdrawal method. Must be bank or momo")
	}

	if len(req.AccountDetails) == 0 {
		return errors.New("account details are required")
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
