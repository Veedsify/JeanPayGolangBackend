package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// GetAdminDashboardStatistics retrieves comprehensive dashboard statistics
func GetAdminDashboardStatistics() (types.DashboardResponse, error) {
	db := database.DB

	var dashboard types.DashboardResponse

	// Get total users count
	var totalUsers int64
	if err := db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		log.Printf("Error getting total users: %v", err)
		return dashboard, err
	}

	// Get total transactions count
	var totalTransactions int64
	if err := db.Model(&models.Transaction{}).Count(&totalTransactions).Error; err != nil {
		log.Printf("Error getting total transactions: %v", err)
		return dashboard, err
	}

	// Get pending transactions count
	var pendingTxns int64
	if err := db.Model(&models.Transaction{}).Where("status = ?", models.TransactionPending).Count(&pendingTxns).Error; err != nil {
		log.Printf("Error getting pending transactions: %v", err)
		return dashboard, err
	}

	// Get completed transactions count
	var completedTxns int64
	if err := db.Model(&models.Transaction{}).Where("status = ?", models.TransactionCompleted).Count(&completedTxns).Error; err != nil {
		log.Printf("Error getting completed transactions: %v", err)
		return dashboard, err
	}

	// Get failed transactions count
	var failedTxns int64
	if err := db.Model(&models.Transaction{}).Where("status = ?", models.TransactionFailed).Count(&failedTxns).Error; err != nil {
		log.Printf("Error getting failed transactions: %v", err)
		return dashboard, err
	}

	// Set summary data
	dashboard.Summary = types.DashboardSummary{
		TotalUsers:        totalUsers,
		TotalTransactions: totalTransactions,
		PendingTxns:       pendingTxns,
		CompletedTxns:     completedTxns,
		FailedTxns:        failedTxns,
	}

	// Get monthly volume by direction
	currentMonth := time.Now().Format("2006-01")
	var monthlyVolumes []types.MonthlyVolume

	// Get deposit volume
	var depositVolume float64
	var depositCount int64
	db.Model(&models.TransactionDetails{}).
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.transaction_type = ? AND DATE_FORMAT(transactions.created_at, '%Y-%m') = ?", models.Deposit, currentMonth).
		Select("COALESCE(SUM(from_amount), 0)").Row().Scan(&depositVolume)
	db.Model(&models.Transaction{}).
		Where("transaction_type = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", models.Deposit, currentMonth).
		Count(&depositCount)

	monthlyVolumes = append(monthlyVolumes, types.MonthlyVolume{
		Direction: "DEPOSIT",
		Total:     depositVolume,
		Count:     depositCount,
	})

	// Get withdrawal volume
	var withdrawalVolume float64
	var withdrawalCount int64
	db.Model(&models.TransactionDetails{}).
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.transaction_type = ? AND DATE_FORMAT(transactions.created_at, '%Y-%m') = ?", models.Withdrawal, currentMonth).
		Select("COALESCE(SUM(from_amount), 0)").Row().Scan(&withdrawalVolume)
	db.Model(&models.Transaction{}).
		Where("transaction_type = ? AND DATE_FORMAT(created_at, '%Y-%m') = ?", models.Withdrawal, currentMonth).
		Count(&withdrawalCount)

	monthlyVolumes = append(monthlyVolumes, types.MonthlyVolume{
		Direction: "WITHDRAWAL",
		Total:     withdrawalVolume,
		Count:     withdrawalCount,
	})

	dashboard.MonthlyVolume = monthlyVolumes

	// Get recent transactions with user info
	var recentTransactions []types.TransactionWithUser
	if err := db.Table("transactions").
		Select("transactions.*, users.first_name, users.last_name, users.email, users.phone_number").
		Joins("LEFT JOIN users ON users.id = transactions.user_id").
		Order("transactions.created_at DESC").
		Limit(10).
		Scan(&recentTransactions).Error; err != nil {
		log.Printf("Error getting recent transactions: %v", err)
		return dashboard, err
	}

	dashboard.RecentTransactions = recentTransactions

	return dashboard, nil
}

// GetAdminUsersAll retrieves paginated list of all users with filters
func GetAdminUsersAll(params types.UserQueryParams) (types.UsersResponse, error) {
	db := database.DB
	var response types.UsersResponse

	// Set default values
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	query := db.Model(&models.User{}).Preload("Setting").Not("is_admin", true)

	// Apply filters
	if params.Status != "" {
		query = query.Where("is_verified = ?", params.Status == "verified")
	}
	if params.Blocked != "" {
		query = query.Where("is_blocked = ?", params.Blocked == "blocked")
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return response, err
	}

	// Get users
	var users []models.User
	if err := query.Offset(offset).Limit(params.Limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return response, err
	}

	// Convert to response format with wallet info
	var userWithWallets []types.UserWithWallet
	for _, user := range users {
		var walletNGN models.Wallet
		var walletGHS models.Wallet
		db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "NGN").First(&walletNGN)
		db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "GHS").First(&walletGHS)

		var transactionCount int64
		db.Model(&models.Transaction{}).Where("user_id = ?", user.ID).Count(&transactionCount)

		wallet := &types.WalletInfo{
			UserID:           user.ID,
			BalanceNGN:       walletNGN.Balance,
			BalanceGHS:       walletGHS.Balance,
			TotalDeposits:    walletNGN.TotalDeposits + walletGHS.TotalDeposits,
			TotalWithdrawals: walletNGN.TotalWithdrawals + walletGHS.TotalWithdrawals,
			TotalConversions: walletNGN.TotalConversions + walletGHS.TotalConversions,
			IsActive:         walletNGN.IsActive || walletGHS.IsActive,
		}

		userWithWallets = append(userWithWallets, types.UserWithWallet{
			User: types.UserInUserwithWallet{
				ID:             user.ID,
				UserID:         user.UserID,
				FirstName:      user.FirstName,
				LastName:       user.LastName,
				Email:          user.Email,
				PhoneNumber:    user.PhoneNumber,
				ProfilePicture: user.ProfilePicture,
				IsVerified:     user.IsVerified,
				IsBlocked:      user.IsBlocked,
				CreatedAt:      user.CreatedAt,
				IsAdmin:        user.IsAdmin,
			},
			Wallet:           wallet,
			TransactionCount: transactionCount,
		})
	}

	response.Users = userWithWallets
	response.Pagination = types.PaginationInfo{
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
		Pages: int((total + int64(params.Limit) - 1) / int64(params.Limit)),
	}

	return response, nil
}

// GetAdminUserDetails retrieves detailed information for a specific user
func GetAdminUserDetails(userID uint) (types.UserWithWallet, error) {
	db := database.DB
	var response types.UserWithWallet

	var user models.User
	if err := db.Preload("Setting").Where("id = ?", userID).First(&user).Error; err != nil {
		return response, err
	}

	var walletNGN models.Wallet
	var walletGHS models.Wallet
	db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "NGN").First(&walletNGN)
	db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "GHS").First(&walletGHS)

	var transactionCount int64
	db.Model(&models.Transaction{}).Where("user_id = ?", user.ID).Count(&transactionCount)

	var recentTransactions []models.Transaction
	if err := db.Model(&models.Transaction{}).
		Preload("TransactionDetails").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentTransactions).Error; err != nil {
		return types.UserWithWallet{}, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	wallet := &types.WalletInfo{
		UserID:           user.ID,
		BalanceNGN:       walletNGN.Balance,
		BalanceGHS:       walletGHS.Balance,
		TotalDeposits:    walletNGN.TotalDeposits + walletGHS.TotalDeposits,
		TotalWithdrawals: walletNGN.TotalWithdrawals + walletGHS.TotalWithdrawals,
		TotalConversions: walletNGN.TotalConversions + walletGHS.TotalConversions,
		WalletIDNGN:      walletNGN.WalletID,
		WalletIDGHS:      walletGHS.WalletID,
		IsActive:         walletNGN.IsActive || walletGHS.IsActive,
	}

	response = types.UserWithWallet{
		User: types.UserInUserwithWallet{
			ID:             user.ID,
			UserID:         user.UserID,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Email:          user.Email,
			Username:       user.Username,
			PhoneNumber:    user.PhoneNumber,
			IsBlocked:      user.IsBlocked,
			CreatedAt:      user.CreatedAt,
			IsVerified:     user.IsVerified,
			IsAdmin:        user.IsAdmin,
			ProfilePicture: user.ProfilePicture, // Assuming ProfilePicture is a field in User model
		},
		Wallet:             wallet,
		TransactionCount:   transactionCount,
		RecentTransactions: recentTransactions,
	}

	return response, nil
}

func UpdateAdminUser(userID uint, data types.UserDetailsEditFields) (types.UserResponse, error) {
	db := database.DB
	var user models.User

	// Find the user by ID
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return types.UserResponse{}, fmt.Errorf("user not found: %w", err)
	}

	// Update user fields
	if strings.Contains(data.Name, " ") {
		parts := strings.Split(data.Name, " ")
		user.FirstName = parts[0]
		if len(parts) > 1 {
			user.LastName = parts[1]
		}
	} else {
		user.FirstName = data.Name
		user.LastName = ""
	}
	user.Email = data.Email
	user.ProfilePicture = data.ProfileImage
	user.Username = data.Username
	user.PhoneNumber = data.Phone

	if data.IsBlocked != nil {
		user.IsBlocked = *data.IsBlocked
	}
	if data.IsVerified != nil {
		user.IsVerified = *data.IsVerified
	}

	// Save changes to the database
	if err := db.Save(&user).Error; err != nil {
		return types.UserResponse{}, fmt.Errorf("failed to update user: %w", err)
	}

	if data.Username != "" {
		// Update username in the setting if it exists
		var setting models.Setting
		if err := db.Where("user_id = ?", user.ID).First(&setting).Error; err == nil {
			setting.Username = data.Username
			if err := db.Save(&setting).Error; err != nil {
				return types.UserResponse{}, fmt.Errorf("failed to update user setting: %w", err)
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return types.UserResponse{}, fmt.Errorf("failed to find user setting: %w", err)
		}
	}

	return types.UserResponse{
		ID:             user.ID,
		UserID:         user.UserID,
		Email:          user.Email,
		ProfilePicture: user.ProfilePicture,
		FirstName:      user.FirstName,
		PhoneNumber:    user.PhoneNumber,
		LastName:       user.LastName,
		Username:       user.Username,
		IsBlocked:      user.IsBlocked,
	}, nil
}

// GetAdminTransactionsAll retrieves paginated list of all transactions
func GetAdminTransactionsAll(params types.AdminTransactionQuery) (types.AdminTransactionResponse, error) {
	db := database.DB
	var response types.AdminTransactionResponse

	// Set default values
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	// Get transactions with preloaded user data and transaction details
	var dbTransactions []models.Transaction
	query := db.Model(&models.Transaction{}).Preload("User").Preload("TransactionDetails")

	// Apply filters
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("transaction_type = ?", params.Type)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("transaction_id LIKE ? OR reference LIKE ?", searchTerm, searchTerm)
	}

	// Get total count
	var total int64
	countQuery := db.Model(&models.Transaction{})
	if params.Status != "" {
		countQuery = countQuery.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		countQuery = countQuery.Where("transaction_type = ?", params.Type)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		countQuery = countQuery.Where("transaction_id LIKE ? OR reference LIKE ?", searchTerm, searchTerm)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return response, err
	}

	// Get transactions
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.Limit).Find(&dbTransactions).Error; err != nil {
		return response, err
	}

	// Convert to response format
	var transactions []types.TransactionWithUser
	for _, tx := range dbTransactions {
		var userInfo *types.UserBasicInfo
		if tx.User.ID != 0 {
			userInfo = &types.UserBasicInfo{
				UserID:      tx.User.ID,
				FirstName:   tx.User.FirstName,
				LastName:    tx.User.LastName,
				Email:       tx.User.Email,
				PhoneNumber: tx.User.PhoneNumber,
			}
		}

		// Get amount and currency from transaction details
		amount := tx.TransactionDetails.FromAmount
		currency := tx.TransactionDetails.FromCurrency

		transactionWithUser := types.TransactionWithUser{
			TransactionID:   tx.TransactionID,
			UserID:          tx.UserID,
			Amount:          amount,
			Currency:        currency,
			Status:          tx.Status,
			TransactionType: tx.TransactionType,
			Reference:       tx.Reference,
			Direction:       tx.Direction,
			Description:     tx.Description,
			CreatedAt:       tx.CreatedAt,
			PaymentType:     tx.PaymentType,
			User:            userInfo,
		}
		transactions = append(transactions, transactionWithUser)
	}

	response.Transactions = transactions
	response.Pagination = types.PaginationInfo{
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
		Pages: int((total + int64(params.Limit) - 1) / int64(params.Limit)),
	}

	return response, nil
}

// GetAdminTransactionDetails retrieves detailed information for a specific transaction
func GetAdminTransactionDetails(transactionID string) (types.AdminTransactionDetailRepsonse, error) {
	db := database.DB
	var transaction models.Transaction

	if err := db.Preload("TransactionDetails").Preload("User").Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
		return types.AdminTransactionDetailRepsonse{}, err
	}

	return types.AdminTransactionDetailRepsonse{
		Transaction: types.TransactionWithUser{
			TransactionID:   transaction.TransactionID,
			UserID:          transaction.UserID,
			TransactionType: models.TransactionType(transaction.TransactionType),
			Reference:       transaction.Reference,
			Status:          transaction.Status,
			Direction:       transaction.Direction,
			Code:            transaction.Code,
			PaymentType:     transaction.PaymentType,
			CreatedAt:       transaction.CreatedAt,
			UpdatedAt:       transaction.UpdatedAt,
			Description:     transaction.Description,
		},
		User: types.UserBasicInfo{
			UserID:         transaction.UserID,
			FirstName:      transaction.User.FirstName,
			LastName:       transaction.User.LastName,
			Email:          transaction.User.Email,
			Username:       transaction.User.Username,
			ProfilePicture: transaction.User.ProfilePicture,
			PhoneNumber:    transaction.User.PhoneNumber,
		},
		TransactionDetails: types.TransactionDetails{
			TransactionID:   transaction.TransactionDetails.TransactionID,
			RecipientName:   transaction.TransactionDetails.RecipientName,
			AccountNumber:   transaction.TransactionDetails.AccountNumber,
			BankName:        transaction.TransactionDetails.BankName,
			PhoneNumber:     transaction.TransactionDetails.PhoneNumber,
			Network:         transaction.TransactionDetails.Network,
			FromCurrency:    transaction.TransactionDetails.FromCurrency,
			ToCurrency:      transaction.TransactionDetails.ToCurrency,
			FromAmount:      transaction.TransactionDetails.FromAmount,
			ToAmount:        transaction.TransactionDetails.ToAmount,
			MethodOfPayment: transaction.TransactionDetails.MethodOfPayment,
		},
	}, nil
}

// ApproveAdminTransaction approves a pending transaction
func ApproveAdminTransaction(transactionID string, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse
	var transaction models.Transaction
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	result := tx.Model(&models.Transaction{}).
		Where("transaction_id = ? AND status = ?", transactionID, models.TransactionPending).
		Updates(map[string]any{
			"status":     models.TransactionCompleted,
			"updated_at": time.Now(),
		})

	// Fetch the updated transaction record to ensure it's returned
	if result.Error == nil && result.RowsAffected > 0 {
		if err := tx.Preload("TransactionDetails").Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
			tx.Rollback()
			return response, err
		}
	}

	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("transaction not found or not in pending status")
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "APPROVE_TRANSACTION",
		Target:   "transaction",
		TargetID: transactionID,
		Details:  fmt.Sprintf("Transaction %s approved", transactionID),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	if transaction.TransactionType == models.Deposit {
		if err := tx.Model(&models.Wallet{}).
			Where("user_id = ? AND currency  = ?", transaction.UserID, transaction.TransactionDetails.FromCurrency).
			Updates(map[string]any{"balance": gorm.Expr("balance + ?", transaction.TransactionDetails.FromAmount)}).Error; err != nil {
			tx.Rollback()
			return response, fmt.Errorf("failed to update wallet balance: %w", err)
		}
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "Transaction approved successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// RejectAdminTransaction rejects a pending transaction
func RejectAdminTransaction(transactionID string, reason string, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	result := tx.Model(&models.Transaction{}).
		Where("transaction_id = ? AND status = ?", transactionID, models.TransactionPending).
		Updates(map[string]any{
			"status":     models.TransactionFailed,
			"reason":     reason,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("transaction not found or not in pending status")
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "REJECT_TRANSACTION",
		Target:   "transaction",
		TargetID: transactionID,
		Details:  fmt.Sprintf("Transaction %s rejected: %s", transactionID, reason),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "Transaction rejected successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetAdminTransactionStatus retrieves transaction status information
func GetAdminTransactionStatus(transactionID string) (map[string]interface{}, error) {
	db := database.DB

	var transaction models.Transaction
	if err := db.Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"transaction_id": transaction.TransactionID,
		"status":         transaction.Status,
		"last_updated":   transaction.UpdatedAt,
		"created_at":     transaction.CreatedAt,
	}

	return status, nil
}

// GetAdminTransactionsOverview retrieves transaction analytics and overview
func GetAdminTransactionsOverview() (types.OverviewSummary, error) {
	db := database.DB
	var overview types.OverviewSummary

	// Calculate summary statistics
	var summary types.OverviewSummary

	// Get total transactions count
	db.Model(&models.Transaction{}).Count(&summary.TotalTransactions)

	// Get pending transactions count
	db.Model(&models.Transaction{}).Where("status = ?", models.TransactionPending).Count(&summary.PendingTransactions)

	// Get completed transactions count
	db.Model(&models.Transaction{}).Where("status = ?", models.TransactionCompleted).Count(&summary.CompletedTransactions)

	// Get failed transactions count
	db.Model(&models.Transaction{}).Where("status = ?", models.TransactionFailed).Count(&summary.FailedTransactions)

	overview = summary

	return overview, nil
}

// GetAdminRatesHistory retrieves exchange rates history
func GetAdminRatesHistory(cursor uint, limit int, searchQuery string) ([]types.RateResponse, uint, error) {
	db := database.DB
	var rates []models.Rate
	query := db.Limit(limit).Order("id DESC").Where("LOWER(from_currency) LIKE ? OR LOWER(to_currency) LIKE ?", "%"+strings.ToLower(searchQuery)+"%", "%"+strings.ToLower(searchQuery)+"%")

	// only apply the cursor if it's not the "first page"
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	if err := query.Find(&rates).Error; err != nil {
		return nil, 0, err
	}

	// convert to response
	var response []types.RateResponse = make([]types.RateResponse, 0, len(rates))
	for _, rate := range rates {
		response = append(response, types.ToRateResponse(&rate))
	}

	// set nextCursor to the last row’s id (because we’re in DESC order)
	nextCursor := uint(0)
	if len(rates) == limit {
		nextCursor = rates[len(rates)-1].ID
	}

	return response, nextCursor, nil
}

// AddAdminRate adds a new exchange rate
func AddAdminRate(rateData types.CreateRateRequest, adminID uint) (types.RateResponse, error) {
	db := database.DB

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new rate
	rate := models.Rate{
		FromCurrency: rateData.FromCurrency,
		ToCurrency:   rateData.ToCurrency,
		Rate:         rateData.Rate,
		Source:       "manual",
		Active:       true,
	}

	if err := tx.Create(&rate).Error; err != nil {
		tx.Rollback()
		return types.RateResponse{}, err
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "ADD_RATE",
		Target:   "rate",
		TargetID: fmt.Sprintf("%d", rate.ID),
		Details:  fmt.Sprintf("Added exchange rate %s to %s: %.4f", rateData.FromCurrency, rateData.ToCurrency, rateData.Rate),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return types.RateResponse{}, err
	}

	tx.Commit()

	return types.ToRateResponse(&rate), nil
}

// BlockUser blocks a user account
func BlockUser(userID string, reason string, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user status

	result := tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_blocked", true)

	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("user not found")
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "BLOCK_USER",
		Target:   "user",
		TargetID: userID,
		Details:  fmt.Sprintf("User %s blocked: %s", userID, reason),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "User blocked successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// UnblockUser unblocks a user account
func UnblockUser(userID string, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user status
	result := tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_blocked", false)

	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("user not found")
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "UNBLOCK_USER",
		Target:   "user",
		TargetID: userID,
		Details:  fmt.Sprintf("User %s unblocked", userID),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "User unblocked successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetUserTransactionHistory retrieves transaction history for a specific user
func GetAdminUserTransactionHistory(userID uint, params types.AdminTransactionQuery) (types.AdminTransactionResponse, error) {
	db := database.DB
	var response types.AdminTransactionResponse

	// Set default values
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	query := db.Table("transactions").
		Select("transactions.*, users.first_name, users.last_name, users.email, users.phone_number").
		Joins("LEFT JOIN users ON users.id = transactions.user_id").
		Where("transactions.user_id = ?", userID)

	// Apply filters
	if params.Status != "" {
		query = query.Where("transactions.status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("transactions.transaction_type = ?", params.Type)
	}

	// Get total count
	var total int64
	db.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&total)

	// Get transactions
	var transactions []types.TransactionWithUser
	if err := query.Order("transactions.created_at DESC").Offset(offset).Limit(params.Limit).Scan(&transactions).Error; err != nil {
		return response, err
	}

	response.Transactions = transactions
	response.Pagination = types.PaginationInfo{
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
		Pages: int((total + int64(params.Limit) - 1) / int64(params.Limit)),
	}

	return response, nil
}

// GetUserWalletDetails retrieves detailed wallet information for a user
func GetUserWalletDetails(userID uint) (*types.WalletInfo, error) {
	db := database.DB

	var walletNGN models.Wallet
	var walletGHS models.Wallet
	db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", userID, "NGN").First(&walletNGN)
	db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", userID, "GHS").First(&walletGHS)

	wallet := &types.WalletInfo{
		UserID:           userID,
		BalanceNGN:       walletNGN.Balance,
		BalanceGHS:       walletGHS.Balance,
		TotalDeposits:    walletNGN.TotalDeposits + walletGHS.TotalDeposits,
		TotalWithdrawals: walletNGN.TotalWithdrawals + walletGHS.TotalWithdrawals,
		TotalConversions: walletNGN.TotalConversions + walletGHS.TotalConversions,
		IsActive:         walletNGN.IsActive || walletGHS.IsActive,
	}

	return wallet, nil
}

// SearchUsers searches for users based on query string
func SearchUsers(query string) ([]types.UserWithWallet, error) {
	db := database.DB
	var users []models.User

	searchTerm := "%" + query + "%"
	if err := db.Preload("Setting").
		Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR phone_number LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm).
		Limit(50).
		Find(&users).Error; err != nil {
		return nil, err
	}

	var userWithWallets []types.UserWithWallet
	for _, user := range users {
		var walletNGN models.Wallet
		var walletGHS models.Wallet
		db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "NGN").First(&walletNGN)
		db.Model(&models.Wallet{}).Where("user_id = ? AND currency = ?", user.ID, "GHS").First(&walletGHS)

		var transactionCount int64
		db.Model(&models.Transaction{}).Where("user_id = ?", user.ID).Count(&transactionCount)

		wallet := &types.WalletInfo{
			UserID:           user.ID,
			BalanceNGN:       walletNGN.Balance,
			BalanceGHS:       walletGHS.Balance,
			TotalDeposits:    walletNGN.TotalDeposits + walletGHS.TotalDeposits,
			TotalWithdrawals: walletNGN.TotalWithdrawals + walletGHS.TotalWithdrawals,
			TotalConversions: walletNGN.TotalConversions + walletGHS.TotalConversions,
			IsActive:         walletNGN.IsActive || walletGHS.IsActive,
		}

		userWithWallets = append(userWithWallets, types.UserWithWallet{
			User: types.UserInUserwithWallet{
				ID:          user.ID,
				UserID:      user.UserID,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				IsBlocked:   user.IsBlocked,
				CreatedAt:   user.CreatedAt,
				IsAdmin:     user.IsAdmin,
			},
			Wallet:           wallet,
			TransactionCount: transactionCount,
		})
	}

	return userWithWallets, nil
}

// UpdateAdminRate updates an existing exchange rate
func UpdateAdminRate(rateID uint, rateData types.UpdateRateRequest, adminID uint) (types.RateResponse, error) {
	db := database.DB

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update rate
	updates := make(map[string]interface{})
	if rateData.Rate != 0 {
		updates["rate"] = rateData.Rate
	}
	if rateData.Source != "" {
		updates["source"] = rateData.Source
	}
	if rateData.Active != nil {
		updates["active"] = *rateData.Active
	}

	var rate models.Rate
	result := tx.Model(&rate).Where("id = ?", rateID).Updates(updates)
	if result.Error != nil {
		tx.Rollback()
		return types.RateResponse{}, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return types.RateResponse{}, fmt.Errorf("rate not found")
	}

	// Get updated rate
	if err := tx.Where("id = ?", rateID).First(&rate).Error; err != nil {
		tx.Rollback()
		return types.RateResponse{}, err
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "UPDATE_RATE",
		Target:   "rate",
		TargetID: fmt.Sprintf("%d", rateID),
		Details:  fmt.Sprintf("Updated exchange rate %s to %s", rate.FromCurrency, rate.ToCurrency),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return types.RateResponse{}, err
	}

	tx.Commit()

	return types.ToRateResponse(&rate), nil
}

// ToggleRateStatus activates or deactivates a rate
func ToggleRateStatus(rateID uint, active bool, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Model(&models.Rate{}).Where("id = ?", rateID).Update("active", active)
	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("rate not found")
	}

	// Log admin action
	action := "DEACTIVATE_RATE"
	if active {
		action = "ACTIVATE_RATE"
	}

	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   action,
		Target:   "rate",
		TargetID: fmt.Sprintf("%d", rateID),
		Details:  fmt.Sprintf("Rate %s", action),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "Rate status updated successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// DeleteRate deletes an exchange rate
func DeleteRate(rateID uint, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get rate details before deletion
	var rate models.Rate
	if err := tx.Where("id = ?", rateID).First(&rate).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	// Delete rate
	result := tx.Delete(&models.Rate{}, rateID)
	if result.Error != nil {
		tx.Rollback()
		return response, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return response, fmt.Errorf("rate not found")
	}

	// Log admin action
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "DELETE_RATE",
		Target:   "rate",
		TargetID: fmt.Sprintf("%d", rateID),
		Details:  fmt.Sprintf("Deleted exchange rate %s to %s", rate.FromCurrency, rate.ToCurrency),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		tx.Rollback()
		return response, err
	}

	tx.Commit()

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "Rate deleted successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetTransactionsByStatus retrieves transactions filtered by status
func GetTransactionsByStatus(status models.TransactionStatus, params types.AdminTransactionQuery) (types.AdminTransactionResponse, error) {
	db := database.DB
	var response types.AdminTransactionResponse

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	query := db.Table("transactions").
		Select("transactions.*, users.first_name, users.last_name, users.email, users.phone_number").
		Joins("LEFT JOIN users ON users.id = transactions.user_id").
		Where("transactions.status = ?", status)

	// Get total count
	var total int64
	db.Model(&models.Transaction{}).Where("status = ?", status).Count(&total)

	// Get transactions
	var transactions []types.TransactionWithUser
	if err := query.Order("transactions.created_at DESC").Offset(offset).Limit(params.Limit).Scan(&transactions).Error; err != nil {
		return response, err
	}

	response.Transactions = transactions
	response.Pagination = types.PaginationInfo{
		Page:  params.Page,
		Limit: params.Limit,
		Total: total,
		Pages: int((total + int64(params.Limit) - 1) / int64(params.Limit)),
	}

	return response, nil
}

// AddTransactionNote adds an admin note to a transaction
func AddTransactionNote(transactionID string, note string, adminID uint) (types.AdminActionResponse, error) {
	db := database.DB
	var response types.AdminActionResponse

	// Verify transaction exists
	var transaction models.Transaction
	if err := db.Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
		return response, fmt.Errorf("transaction not found")
	}

	// Log admin action with note
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "ADD_NOTE",
		Target:   "transaction",
		TargetID: transactionID,
		Details:  fmt.Sprintf("Added note: %s", note),
	}
	if err := db.Create(&adminLog).Error; err != nil {
		return response, err
	}

	response = types.AdminActionResponse{
		Success:   true,
		Message:   "Note added successfully",
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetUserActivityLogs retrieves activity logs for a specific user
func GetUserActivityLogs(userID string) ([]models.Activity, error) {
	db := database.DB
	var activities []models.Activity

	if err := db.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&activities).Error; err != nil {
		return nil, err
	}

	return activities, nil
}
