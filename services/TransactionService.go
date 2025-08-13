package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/jobs"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"gorm.io/gorm"
)

func CreateTransaction(userId uint, transaction types.NewTransactionRequest) (types.CreateNewTransactionResponse, string, error) {
	if userId == 0 {
		return types.CreateNewTransactionResponse{}, "", errors.New("user is not valid")
	}
	user, err := GetUserById(userId)
	balances, err2 := GetWalletBalance(userId)
	if err != nil {
		return types.CreateNewTransactionResponse{}, "", errors.New("user not found")
	}
	if err2 != nil {
		return types.CreateNewTransactionResponse{}, "", errors.New("wallet balance not found")
	}
	TransactionIdx, err := libs.SecureRandomNumber(16)
	if err != nil {
		return types.CreateNewTransactionResponse{}, "INTERNAL_SERVER_ERROR", errors.New("failed to generate transaction index")
	}
	var transactionDir = utils.GetConvertdirection(transaction.FromCurrency)
	notificationClient := jobs.NewNotificationJobClient()

	switch transaction.MethodOfPayment {
	case "wallet":
		fromAmount, err := utils.ConvertStringToFloat(transaction.FromAmount)
		if err != nil {
			return types.CreateNewTransactionResponse{}, "INVALID_AMOUNTS", errors.New("invalid transaction amounts")
		}
		toAmount, err := utils.ConvertStringToFloat(transaction.ToAmount)
		if err != nil {
			return types.CreateNewTransactionResponse{}, "INVALID_AMOUNTS", errors.New("invalid amount")
		}
		for _, wallet := range balances {
			if wallet.Currency == transaction.FromCurrency {
				if wallet.Balance < fromAmount {
					return types.CreateNewTransactionResponse{}, "INSUFFICIENT_FUNDS", errors.New("insufficient wallet balance")
				}
				code, err := HandleWalletTransaction(wallet, transaction)
				if err != nil {
					return types.CreateNewTransactionResponse{}, code, err
				}
			}
		}
		transactionId := fmt.Sprintf("TRX%d", TransactionIdx)
		transaction := models.Transaction{
			UserID:          user.ID,
			TransactionID:   transactionId,
			PaymentType:     models.PaymentType(transaction.Method),
			Status:          models.TransactionPending,
			TransactionType: models.Transfer,
			Description:     fmt.Sprintf("Transfer %s %s to %s", transaction.ToCurrency, transaction.ToAmount, transaction.RecipientName),
			Reference:       libs.GenerateUniqueID(),
			Direction:       transactionDir,
			TransactionDetails: models.TransactionDetails{
				ToCurrency:      transaction.ToCurrency,
				FromCurrency:    transaction.FromCurrency,
				FromAmount:      fromAmount,
				ToAmount:        toAmount,
				RecipientName:   transaction.RecipientName,
				AccountNumber:   transaction.AccountNumber,
				BankName:        transaction.BankName,
				PhoneNumber:     transaction.PhoneNumber,
				Network:         transaction.Network,
				MethodOfPayment: transaction.MethodOfPayment,
			},
		}
		if err := database.DB.Create(&transaction).Error; err != nil {
			return types.CreateNewTransactionResponse{}, "INTERNAL_SERVER_ERROR", errors.New("failed to create transaction")
		}

		title := "Transaction Successful"
		message := fmt.Sprintf("Your transfer of %s %s to %s was successful.", transaction.TransactionDetails.FromCurrency, transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.RecipientName)

		notificationClient.EnqueueCreateNotification(
			user.ID,
			models.NotificationType("transfer"),
			title,
			message,
		)

		return types.CreateNewTransactionResponse{
			Transaction: types.TransactionResponse{
				ID:              transaction.ID,
				TransactionID:   transaction.TransactionID,
				UserID:          transaction.UserID,
				Status:          transaction.Status,
				TransactionType: transaction.TransactionType,
				Reference:       transaction.Reference,
				Direction:       transaction.Direction,
				Description:     transaction.Description,
				CreatedAt:       transaction.CreatedAt,
				UpdatedAt:       transaction.UpdatedAt,
			},
			RedirectionURL: "",
			ShouldRedirect: false,
		}, "", nil
	case "checkout":
		transactionId := fmt.Sprintf("TRX%d", TransactionIdx)
		response, code, err := HandleDirectTransaction(transaction, transactionId, user.ID)
		if err != nil {
			return types.CreateNewTransactionResponse{
				Transaction:    types.TransactionResponse{},
				RedirectionURL: "",
				ShouldRedirect: false,
			}, code, err
		}
		return response, "", nil
	}
	return types.CreateNewTransactionResponse{}, "", errors.New("invalid payment method")
}
func HandleDirectTransaction(transaction types.NewTransactionRequest, transactionId string, userId uint) (types.CreateNewTransactionResponse, string, error) {
	if transaction.FromAmount == "" || transaction.ToAmount == "" {
		return types.CreateNewTransactionResponse{}, "INVALID_AMOUNT", errors.New("invalid transaction amounts")
	}

	fromAmount, err := utils.ConvertStringToFloat(transaction.FromAmount)
	if err != nil {
		return types.CreateNewTransactionResponse{}, "INVALID_AMOUNTS", errors.New("invalid transaction amounts")
	}
	toAmount, err := utils.ConvertStringToFloat(transaction.ToAmount)
	if err != nil {
		return types.CreateNewTransactionResponse{}, "INVALID_AMOUNTS", errors.New("invalid amount")
	}

	var transactionDir = utils.GetConvertdirection(transaction.FromCurrency)

	// Create a pending transaction for direct payment
	pendingTransaction := models.Transaction{
		UserID:          userId,
		TransactionID:   transactionId,
		PaymentType:     models.PaymentType(transaction.Method),
		Status:          models.TransactionPending,
		TransactionType: models.Transfer,
		Reference:       libs.GenerateUniqueID(),
		Direction:       transactionDir,
		Description:     fmt.Sprintf("Transfer %s %s to %s", transaction.ToCurrency, transaction.ToAmount, transaction.RecipientName),
		TransactionDetails: models.TransactionDetails{
			ToCurrency:      transaction.ToCurrency,
			FromCurrency:    transaction.FromCurrency,
			FromAmount:      fromAmount,
			ToAmount:        toAmount,
			RecipientName:   transaction.RecipientName,
			AccountNumber:   transaction.AccountNumber,
			BankName:        transaction.BankName,
			PhoneNumber:     transaction.PhoneNumber,
			Network:         transaction.Network,
			MethodOfPayment: transaction.MethodOfPayment,
		},
	}

	if err := database.DB.Create(&pendingTransaction).Error; err != nil {
		return types.CreateNewTransactionResponse{}, "INTERNAL_SERVER_ERROR", errors.New("failed to create transaction")
	}

	return types.CreateNewTransactionResponse{
		Transaction: types.TransactionResponse{
			ID:              pendingTransaction.ID,
			TransactionID:   pendingTransaction.TransactionID,
			UserID:          pendingTransaction.UserID,
			Status:          pendingTransaction.Status,
			TransactionType: pendingTransaction.TransactionType,
			Reference:       pendingTransaction.Reference,
			Direction:       pendingTransaction.Direction,
			Description:     pendingTransaction.Description,
			CreatedAt:       pendingTransaction.CreatedAt,
			UpdatedAt:       pendingTransaction.UpdatedAt,
		},
		RedirectionURL: "",
		ShouldRedirect: false,
	}, "", nil
}
func HandleWalletTransaction(wallet types.WalletBalance, transaction types.NewTransactionRequest) (string, error) {
	fromAmount, err := utils.ConvertStringToFloat(transaction.FromAmount)
	if err != nil {
		return "INVALID_AMOUNT", errors.New("invalid from amount")
	}
	if wallet.Balance < fromAmount && wallet.Currency == transaction.FromCurrency {
		return "INSUFFICIENT_FUNDS", errors.New("insufficient wallet balance")
	}
	code, err := HandleRemoveMoneyFromWallet(wallet, fromAmount)
	if err != nil {
		return code, err
	}
	return "", nil
}
func HandleRemoveMoneyFromWallet(wallet types.WalletBalance, amount float64) (string, error) {
	if wallet.Balance < amount {
		return "INSUFFICIENT_FUNDS", errors.New("insufficient wallet balance")
	}
	var userWallet models.Wallet
	if err := database.DB.First(&userWallet, wallet.ID).Error; err != nil {
		return "INTERNAL_SERVER_ERROR", errors.New("failed to fetch wallet")
	}
	if userWallet.Balance < amount {
		return "INSUFFICIENT_FUNDS", errors.New("insufficient wallet balance")
	}
	userWallet.Balance -= amount
	if err := database.DB.Save(&userWallet).Error; err != nil {
		return "INTERNAL_SERVER_ERROR", errors.New("failed to update wallet")
	}
	return "", nil
}

// GetAllTransactions retrieves all transactions with filtering and pagination (Admin only)
func GetAllTransactions(filter types.TransactionFilter, pagination types.PaginationRequest) ([]types.TransactionResponse, *types.PaginationResponse, error) {
	// Build a query
	query := database.DB.Model(&models.Transaction{})
	// Apply filters
	if filter.Status != "" && isValidTransactionStatus(filter.Status) {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" && isValidTransactionType(filter.Type) {
		query = query.Where("transaction_type = ?", filter.Type)
	}
	if filter.Currency != "" && isValidCurrency(filter.Currency) {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.FromDate != "" && filter.ToDate != "" {
		fromDate, err := time.Parse("2006-01-02", filter.FromDate)
		if err == nil {
			toDate, err := time.Parse("2006-01-02", filter.ToDate)
			if err == nil {
				query = query.Where("created_at >= ? AND created_at <= ?", fromDate, toDate.Add(24*time.Hour-time.Second))
			}
		}
	}
	if filter.Search != "" {
		query = query.Where("transaction_id LIKE ? OR reference LIKE ? OR description LIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
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
	// Find transactions with user data
	var transactions []models.Transaction
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find transactions: %w", err)
	}
	// Convert to response format
	var transactionResponses []types.TransactionResponse
	for _, txn := range transactions {
		response := types.TransactionResponse{
			ID:              txn.ID,
			TransactionID:   txn.TransactionID,
			UserID:          txn.UserID,
			Status:          txn.Status,
			TransactionType: txn.TransactionType,
			Reference:       txn.Reference,
			Direction:       txn.Direction,
			Description:     txn.Description,
			CreatedAt:       txn.CreatedAt,
			UpdatedAt:       txn.UpdatedAt,
		}
		// Add user info if available
		if txn.User.ID != 0 {
			response.User = &types.UserInfo{
				FirstName: txn.User.FirstName,
				LastName:  txn.User.LastName,
				Email:     txn.User.Email,
			}
		}
		transactionResponses = append(transactionResponses, response)
	}
	// Create pagination response
	paginationResp := &types.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}
	return transactionResponses, paginationResp, nil
}

// GetUserTransactionHistory retrieves transaction history for a specific user
func GetUserTransactionHistory(userID uint32, filter types.TransactionFilter, pagination types.PaginationRequest) ([]types.TransactionResponse, *types.PaginationResponse, error) {
	if userID == 0 {
		return nil, nil, errors.New("user ID is required")
	}
	// Build query for user transactions
	query := database.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)
	// Apply filters
	if filter.Status != "" && isValidTransactionStatus(filter.Status) {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" && isValidTransactionType(filter.Type) {
		query = query.Where("transaction_type = ?", filter.Type)
	}
	if filter.FromDate != "" && filter.ToDate != "" {
		fromDate, err := time.Parse("2006-01-02", filter.FromDate)
		if err == nil {
			toDate, err := time.Parse("2006-01-02", filter.ToDate)
			if err == nil {
				query = query.Where("created_at >= ? AND created_at <= ?", fromDate, toDate.Add(24*time.Hour-time.Second))
			}
		}
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
		limit = 10
	}
	offset := (page - 1) * limit
	// Find transactions
	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find transactions: %w", err)
	}
	// Convert to response format
	var transactionResponses []types.TransactionResponse
	for _, txn := range transactions {
		transactionResponses = append(transactionResponses, types.TransactionResponse{
			ID:              txn.ID,
			TransactionID:   txn.TransactionID,
			UserID:          txn.UserID,
			Status:          txn.Status,
			TransactionType: txn.TransactionType,
			Reference:       txn.Reference,
			Direction:       txn.Direction,
			Description:     txn.Description,
			CreatedAt:       txn.CreatedAt,
			UpdatedAt:       txn.UpdatedAt,
		})
	}
	// Create pagination response
	paginationResp := &types.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}
	return transactionResponses, paginationResp, nil
}

// GetTransactionDetails retrieves detailed information about a specific transaction
func GetTransactionDetails(transactionID string, userID uint32) (*types.TransactionResponse, error) {
	if transactionID == "" {
		return nil, errors.New("transaction ID is required")
	}
	var transaction models.Transaction
	query := database.DB.Preload("User")
	// If userID is provided, ensure users can only see their own transactions
	if userID != 0 {
		query = query.Where("transaction_id = ? AND user_id = ?", transactionID, userID)
	} else {
		query = query.Where("transaction_id = ?", transactionID)
	}
	if err := query.First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	response := &types.TransactionResponse{
		ID:              transaction.ID,
		TransactionID:   transaction.TransactionID,
		UserID:          transaction.UserID,
		Status:          transaction.Status,
		TransactionType: transaction.TransactionType,
		Reference:       transaction.Reference,
		Direction:       transaction.Direction,
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}
	// Add user info if available
	if transaction.User.ID != 0 {
		response.User = &types.UserInfo{
			FirstName: transaction.User.FirstName,
			LastName:  transaction.User.LastName,
			Email:     transaction.User.Email,
		}
	}
	return response, nil
}

// UpdateTransactionStatus updates the status of a transaction (Admin only)
func UpdateTransactionStatus(transactionID string, newStatus string, adminID uint32, reason string) error {
	if transactionID == "" {
		return errors.New("transaction ID is required")
	}
	if !isValidTransactionStatus(newStatus) {
		return errors.New("invalid transaction status")
	}
	if adminID == 0 {
		return errors.New("admin ID is required")
	}
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("transaction not found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}
	// Update transaction
	updates := map[string]interface{}{
		"status":     newStatus,
		"updated_at": time.Now(),
	}
	if reason != "" {
		updates["description"] = transaction.Description + " | Admin note: " + reason
	}
	if err := database.DB.Model(&transaction).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	// Create admin log entry
	adminLog := models.AdminLog{
		AdminID:  adminID,
		Action:   "update_transaction_status",
		Target:   "transaction",
		TargetID: transactionID,
		Details:  fmt.Sprintf("Status changed to %s. Reason: %s", newStatus, reason),
	}
	if err := database.DB.Create(&adminLog).Error; err != nil {
		// Log error but don't fail the transaction update
		fmt.Printf("Failed to create admin log: %v\n", err)
	}
	return nil
}

// GetTransactionByReference finds a transaction by its reference
func GetTransactionByReference(reference string) (*models.Transaction, error) {
	if reference == "" {
		return nil, errors.New("reference is required")
	}
	var transaction models.Transaction
	if err := database.DB.Where("reference = ?", reference).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	return &transaction, nil
}

// UpdateTransactionStatusByReference updates transaction status using reference
func UpdateTransactionStatusByReference(reference string, status string) error {
	if reference == "" {
		return errors.New("reference is required")
	}
	if !isValidTransactionStatus(status) {
		return errors.New("invalid transaction status")
	}
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now(),
	}
	result := database.DB.Model(&models.Transaction{}).Where("reference = ?", reference).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("transaction not found")
	}
	return nil
}

// GetTransactionStats gets transaction statistics
func GetTransactionStats(userID uint32, period string) (map[string]interface{}, error) {
	var dateFilter time.Time
	switch period {
	case "today":
		dateFilter = time.Now().Truncate(24 * time.Hour)
	case "week":
		dateFilter = time.Now().AddDate(0, 0, -7)
	case "month":
		dateFilter = time.Now().AddDate(0, -1, 0)
	case "year":
		dateFilter = time.Now().AddDate(-1, 0, 0)
	default:
		dateFilter = time.Now().AddDate(0, -1, 0) // Default to last month
	}
	query := database.DB.Model(&models.Transaction{}).Where("created_at >= ?", dateFilter)
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	var totalCount, completedCount, pendingCount, failedCount int64
	var totalVolume, completedVolume float64
	// Get counts
	query.Count(&totalCount)
	query.Where("status = ?", "completed").Count(&completedCount)
	query.Where("status = ?", "pending").Count(&pendingCount)
	query.Where("status = ?", "failed").Count(&failedCount)
	// Get volumes
	var volumeResult struct {
		Total float64
	}
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("created_at >= ?", dateFilter).
		Scan(&volumeResult)
	totalVolume = volumeResult.Total
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("status = ? AND created_at >= ?", "completed", dateFilter).
		Scan(&volumeResult)
	completedVolume = volumeResult.Total
	if userID != 0 {
		database.DB.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0) as total").
			Where("user_id = ? AND created_at >= ?", userID, dateFilter).
			Scan(&volumeResult)
		totalVolume = volumeResult.Total
		database.DB.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0) as total").
			Where("user_id = ? AND status = ? AND created_at >= ?", userID, "completed", dateFilter).
			Scan(&volumeResult)
		completedVolume = volumeResult.Total
	}
	stats := map[string]any{
		"period": period,
		"counts": map[string]int64{
			"total":     totalCount,
			"completed": completedCount,
			"pending":   pendingCount,
			"failed":    failedCount,
		},
		"volumes": map[string]float64{
			"total":     libs.RoundCurrency(totalVolume),
			"completed": libs.RoundCurrency(completedVolume),
		},
	}
	return stats, nil
}

// GetUserTransactionHistoryService retrieves transaction history for a specific user
func GetUserTransactionHistoryService(userID uint, query types.UserTransactionQuery) (*types.AdminTransactionResponse, error) {
	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	offset := (query.Page - 1) * query.Limit
	// Build a query
	db := database.DB.Model(&models.Transaction{}).Where("user_id = ?", userID).Preload("User").Preload("TransactionDetails")
	// Apply filters
	if query.Status != "" && isValidTransactionStatus(query.Status) {
		db = db.Where("status = ?", query.Status)
	}
	if query.Type != "" && isValidTransactionType(query.Type) {
		db = db.Where("transaction_type = ?", query.Type)
	}
	if query.Currency != "" && isValidCurrency(query.Currency) {
		db = db.Where("currency = ?", query.Currency)
	}
	if query.AccountType != "" {
		db = db.Where("payment_type = ?", query.AccountType)
	}
	if query.Search != "" {
		like := "%" + query.Search + "%"
		db = db.Where("transaction_id LIKE ? OR reference LIKE ? OR description LIKE ?", like, like, like)
	}
	if query.FromDate != "" && query.ToDate != "" {
		fromDate, err := time.Parse("2006-01-02", query.FromDate)
		if err == nil {
			toDate, err := time.Parse("2006-01-02", query.ToDate)
			if err == nil {
				db = db.Where("created_at >= ? AND created_at <= ?", fromDate, toDate.Add(24*time.Hour-time.Second))
			}
		}
	}
	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("failed to count transactions")
	}
	// Get transactions
	var transactions []models.Transaction
	err := db.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&transactions).Error
	if err != nil {
		return nil, errors.New("failed to get transactions")
	}
	// Convert to response format
	var transactionsWithUsers []types.TransactionWithUser
	for _, transaction := range transactions {
		transactionsWithUsers = append(transactionsWithUsers, types.TransactionWithUser{
			TransactionID:      transaction.TransactionID,
			UserID:             transaction.UserID,
			TransactionType:    string(transaction.TransactionType),
			Status:             string(transaction.Status),
			Reference:          transaction.Reference,
			Direction:          string(transaction.Direction),
			Description:        transaction.Description,
			CreatedAt:          transaction.CreatedAt,
			PaymentType:        string(transaction.PaymentType),
			TransactionDetails: transaction.TransactionDetails,
			User: &types.UserBasicInfo{
				UserID:    transaction.UserID,
				FirstName: transaction.User.FirstName,
				LastName:  transaction.User.LastName,
				Email:     transaction.User.Email,
			},
		})
	}
	response := &types.AdminTransactionResponse{
		Transactions: transactionsWithUsers,
		Pagination: types.PaginationInfo{
			Page:  query.Page,
			Limit: query.Limit,
			Total: total,
			Pages: int(math.Ceil(float64(total) / float64(query.Limit))),
		},
	}
	return response, nil
}

// GetTransactionDetailsService retrieves detailed transaction information
func GetTransactionDetailsService(transactionID string, userID uint, isAdmin bool) (*types.TransactionDetailsResponse, error) {
	var transaction models.Transaction
	if isAdmin {
		// Admins can view any transaction
		err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("transaction not found")
			}
			return nil, errors.New("failed to get transaction")
		}
	} else {
		// Users can only view their own transactions
		err := database.DB.Preload("TransactionDetails").Where("transaction_id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("transaction not found")
			}
			return nil, errors.New("failed to get transaction")
		}
	}
	// Get user details
	var user models.User
	err := database.DB.First(&user, transaction.UserID).Error
	if err != nil {
		return nil, errors.New("failed to get user details")
	}
	response := &types.TransactionDetailsResponse{
		Transaction: transaction,
	}
	return response, nil
}

// UpdateTransactionStatusService updates transaction status
func UpdateTransactionStatusService(transactionID string, request types.UpdateTransactionStatusRequest, adminID uint32) (*types.TransactionActionResponse, error) {
	var transaction models.Transaction
	err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, errors.New("failed to get transaction")
	}
	// Start transaction
	tx := database.DB.Begin()
	// Update transaction status
	oldStatus := transaction.Status
	transaction.Status = models.TransactionStatus(request.Status)
	if request.Reason != "" {
		transaction.Description += fmt.Sprintf(" | Admin update: %s", request.Reason)
	}
	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update transaction status")
	}
	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  adminID,
		Action:   "update_transaction_status",
		Target:   "transaction",
		TargetID: transactionID,
		Details: fmt.Sprint(map[string]interface{}{
			"transaction_id":  transactionID,
			"previous_status": string(oldStatus),
			"new_status":      request.Status,
			"reason":          request.Reason,
		}),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to log admin action")
	}
	tx.Commit()
	response := &types.TransactionActionResponse{
		TransactionID: transactionID,
		Status:        request.Status,
		ActionedAt:    time.Now().Format(time.RFC3339),
		ActionedBy:    adminID,
		Reason:        request.Reason,
	}
	return response, nil
}

// CreateDepositTransactionService creates a new deposit transaction
func CreateDepositTransactionService(userID uint, request types.CreateDepositRequest) (*types.TransactionWithUser, error) {
	// Validate currency
	if !isValidCurrency(request.Currency) {
		return nil, errors.New("invalid currency")
	}
	if request.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	// Generate transaction ID and reference
	transactionID := libs.GenerateUniqueID()
	reference := fmt.Sprintf("DEP_%s_%d", request.Currency, time.Now().Unix())
	// Determine direction
	var direction models.TransactionDirection
	if request.Currency == "NGN" {
		direction = models.DepositNGN
	} else {
		direction = models.DepositGHS
	}
	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		Status:          models.TransactionPending,
		TransactionType: models.Deposit,
		Reference:       reference,
		Direction:       direction,
		Description:     request.Description,
	}
	if err := database.DB.Create(&transaction).Error; err != nil {
		return nil, errors.New("failed to create deposit transaction")
	}
	response := &types.TransactionWithUser{
		TransactionID:   transaction.TransactionID,
		UserID:          transaction.UserID,
		Status:          string(transaction.Status),
		TransactionType: string(transaction.TransactionType),
		Reference:       transaction.Reference,
		Direction:       string(transaction.Direction),
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
	}
	return response, nil
}

// CreateWithdrawalTransactionService creates a new withdrawal transaction
func CreateWithdrawalTransactionService(userID uint, request types.CreateWithdrawalRequest) (*types.TransactionWithUser, error) {
	// Validate currency
	if !isValidCurrency(request.Currency) {
		return nil, errors.New("invalid currency")
	}
	if request.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	// Check wallet balance
	var wallet models.Wallet
	err := database.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, errors.New("wallet not found")
	}
	var balance float64
	if request.Currency == "NGN" {
		balance = wallet.Balance
	} else {
		balance = wallet.Balance
	}
	if balance < request.Amount {
		return nil, errors.New("insufficient balance")
	}
	// Start transaction
	tx := database.DB.Begin()
	// Deduct from wallet balance
	if request.Currency == "NGN" {
		wallet.Balance -= request.Amount
	} else {
		wallet.Balance -= request.Amount
	}
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update wallet balance")
	}
	// Generate transaction ID and reference
	transactionID := libs.GenerateUniqueID()
	reference := fmt.Sprintf("WDR_%s_%d", request.Currency, time.Now().Unix())
	// Determine direction
	var direction models.TransactionDirection
	if request.Currency == "NGN" {
		direction = models.WithdrawalNGN
	} else {
		direction = models.WithdrawalGHS
	}
	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		Status:          models.TransactionPending,
		TransactionType: models.Withdrawal,
		Reference:       reference,
		Direction:       direction,
		Description:     fmt.Sprintf("Withdrawal to %s. %s", request.BankAccount, request.Description),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create withdrawal transaction")
	}
	tx.Commit()
	response := &types.TransactionWithUser{
		TransactionID:   transaction.TransactionID,
		UserID:          transaction.UserID,
		TransactionType: string(transaction.TransactionType),
		Status:          string(transaction.Status),
		Reference:       transaction.Reference,
		Direction:       string(transaction.Direction),
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
		User: &types.UserBasicInfo{
			Email:     transaction.User.Email,
			FirstName: transaction.User.FirstName,
			LastName:  transaction.User.LastName,
			UserID:    transaction.User.ID,
		},
	}
	return response, nil
}

// GetTransactionStatsService returns transaction statistics for users
func GetTransactionStatsService(userID string) (*types.TransactionStatsResponse, error) {
	currentTime := time.Now()
	startOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())
	var stats types.TransactionStatsResponse
	// Get total counts
	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND status = ?", userID, models.TransactionPending).Count(&stats.PendingCount)
	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND status = ?", userID, models.TransactionCompleted).Count(&stats.CompletedCount)
	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND status = ?", userID, models.TransactionFailed).Count(&stats.FailedCount)
	// Get total amounts by type
	var depositResult, withdrawalResult, conversionResult struct {
		Total float64
	}
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, models.Deposit, models.TransactionCompleted).
		Scan(&depositResult)
	stats.TotalDeposits = depositResult.Total
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, models.Withdrawal, models.TransactionCompleted).
		Scan(&withdrawalResult)
	stats.TotalWithdrawals = withdrawalResult.Total
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, models.Conversion, models.TransactionCompleted).
		Scan(&conversionResult)
	stats.TotalConversions = conversionResult.Total
	// Get monthly amounts
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_type = ? AND status = ? AND created_at >= ?", userID, models.Deposit, models.TransactionCompleted, startOfMonth).
		Scan(&depositResult)
	stats.MonthlyDeposits = depositResult.Total
	database.DB.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND transaction_type = ? AND status = ? AND created_at >= ?", userID, models.Withdrawal, models.TransactionCompleted, startOfMonth).
		Scan(&withdrawalResult)
	stats.MonthlyWithdrawals = withdrawalResult.Total
	return &stats, nil
}

// Helper functions
// isValidCurrency checks if currency is valid
func isValidCurrency(currency string) bool {
	validCurrencies := []string{"NGN", "GHS"}
	for _, c := range validCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// FilterTransactions filters transactions based on criteria
func FilterTransactions(userID uint32, filterReq types.TransactionFilterRequest) (*types.GetTransactionsResponse, error) {
	query := database.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)
	// Apply filters
	if filterReq.Status != "" {
		query = query.Where("status = ?", filterReq.Status)
	}
	if filterReq.Type != "" {
		query = query.Where("transaction_type = ?", filterReq.Type)
	}
	if filterReq.Currency != "" {
		query = query.Where("currency = ?", filterReq.Currency)
	}
	if filterReq.Direction != "" {
		query = query.Where("direction = ?", filterReq.Direction)
	}
	if filterReq.MinAmount > 0 {
		query = query.Where("amount >= ?", filterReq.MinAmount)
	}
	if filterReq.MaxAmount > 0 {
		query = query.Where("amount <= ?", filterReq.MaxAmount)
	}
	if !filterReq.FromDate.IsZero() {
		query = query.Where("created_at >= ?", filterReq.FromDate)
	}
	if !filterReq.ToDate.IsZero() {
		query = query.Where("created_at <= ?", filterReq.ToDate)
	}
	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count filtered transactions: %w", err)
	}
	// Pagination
	page := filterReq.Page
	if page <= 0 {
		page = 1
	}
	limit := filterReq.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	// Get transactions
	var transactions []models.Transaction
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get filtered transactions: %w", err)
	}
	// Convert to response format
	responseTransactions := types.ToTransactionsResponse(transactions)
	return &types.GetTransactionsResponse{
		Transactions: responseTransactions,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   int((total + int64(limit) - 1) / int64(limit)),
	}, nil
}
