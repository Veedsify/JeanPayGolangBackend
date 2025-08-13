package services

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// AdminLogin handles admin authentication
func AdminLogin(req types.AdminLoginRequest) (*types.AdminLoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Find admin user
	var admin models.User
	err := database.DB.Where("email = ? AND is_admin = ?", req.Email, true).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid admin credentials")
		}
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}

	// Check if admin is verified
	if !admin.IsVerified {
		return nil, errors.New("admin account is not verified")
	}

	// Verify password
	if err := libs.ComparePassword(admin.Password, req.Password); err != nil {
		return nil, errors.New("invalid admin credentials")
	}

	// Generate token
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		return nil, err
	}

	userInfo := &libs.UserInfo{
		ID:      admin.ID,
		UserID:  admin.UserID,
		Email:   admin.Email,
		IsAdmin: admin.IsAdmin,
	}

	tokenPair, err := jwtService.GenerateTokenPair(userInfo)
	if err != nil {
		return nil, err
	}

	response := &types.AdminLoginResponse{
		Token: tokenPair.AccessToken,
		Admin: types.AdminProfileResponse{
			UserID:    admin.UserID,
			FirstName: admin.FirstName,
			LastName:  admin.LastName,
			Email:     admin.Email,
			IsAdmin:   admin.IsAdmin,
		},
	}

	return response, nil
}

// GetAdminDashboard returns admin dashboard data
func GetAdminDashboard() (*types.DashboardResponse, error) {
	// Get summary statistics
	var summary types.DashboardSummary

	database.DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&summary.TotalUsers)
	database.DB.Model(&models.Transaction{}).Count(&summary.TotalTransactions)
	database.DB.Model(&models.Transaction{}).Where("status = ?", models.TransactionPending).Count(&summary.PendingTxns)
	database.DB.Model(&models.Transaction{}).Where("status = ?", models.TransactionCompleted).Count(&summary.CompletedTxns)
	database.DB.Model(&models.Transaction{}).Where("status = ?", models.TransactionFailed).Count(&summary.FailedTxns)

	// Get monthly volume data
	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())

	var monthlyVolume []types.MonthlyVolume
	directions := []string{"NGN-GHS", "GHS-NGN", "DEPOSIT-NGN", "DEPOSIT-GHS", "WITHDRAWAL-NGN", "WITHDRAWAL-GHS"}

	for _, direction := range directions {
		var result struct {
			Total float64
			Count int64
		}

		database.DB.Model(&models.Transaction{}).
			Select("COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
			Where("direction = ? AND created_at >= ? AND status = ?", direction, startOfMonth, models.TransactionCompleted).
			Scan(&result)

		if result.Count > 0 {
			monthlyVolume = append(monthlyVolume, types.MonthlyVolume{
				Direction: direction,
				Total:     libs.RoundCurrency(result.Total),
				Count:     result.Count,
			})
		}
	}

	// Get recent transactions
	var transactions []models.Transaction
	database.DB.Order("created_at DESC").Limit(10).Find(&transactions)

	var recentTransactions []types.TransactionWithUser
	for _, txn := range transactions {
		var user models.User
		database.DB.Where("user_id = ?", txn.UserID).First(&user)

		userInfo := &types.UserBasicInfo{
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
		}

		recentTransactions = append(recentTransactions, types.TransactionWithUser{
			TransactionID:   txn.TransactionID,
			UserID:          txn.UserID,
			TransactionType: string(txn.TransactionType),
			Status:          string(txn.Status),
			Reference:       txn.Reference,
			Direction:       string(txn.Direction),
			Description:     string(txn.Description),
			CreatedAt:       txn.CreatedAt,
			User:            userInfo,
		})
	}

	return &types.DashboardResponse{
		Summary:            summary,
		MonthlyVolume:      monthlyVolume,
		RecentTransactions: recentTransactions,
	}, nil
}

// GetAllTransactionsAdminService returns paginated transactions for admin
func GetAllTransactionsAdminService(query types.TransactionQueryParams) (*types.AdminTransactionResponse, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}

	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filter conditions
	tx := database.DB.Model(&models.Transaction{})

	if query.Status != "" && isValidTransactionStatus(query.Status) {
		tx = tx.Where("status = ?", query.Status)
	}

	if query.Type != "" && isValidTransactionType(query.Type) {
		tx = tx.Where("transaction_type = ?", query.Type)
	}

	if query.Search != "" {
		// Search by transaction ID, user email, or reference
		var userIDs []string
		database.DB.Model(&models.User{}).
			Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
				"%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%").
			Pluck("user_id", &userIDs)

		tx = tx.Where("transaction_id ILIKE ? OR reference ILIKE ? OR user_id IN ?",
			"%"+query.Search+"%", "%"+query.Search+"%", userIDs)
	}

	// Get total count
	var total int64
	tx.Count(&total)

	// Get transactions
	var transactions []models.Transaction
	tx.Order("created_at desc").Offset(offset).Limit(limit).Find(&transactions)

	// Get user details for each transaction
	transactionsWithUsers := make([]types.TransactionWithUser, len(transactions))
	for i, transaction := range transactions {
		var user models.User
		database.DB.Select("first_name, last_name, email, phone_number").
			Where("user_id = ?", transaction.UserID).First(&user)

		transactionsWithUsers[i] = types.TransactionWithUser{
			TransactionID:   transaction.TransactionID,
			UserID:          transaction.UserID,
			TransactionType: string(transaction.TransactionType),
			Status:          string(transaction.Status),
			Reference:       transaction.Reference,
			Direction:       string(transaction.Direction),
			Description:     string(transaction.Description),
			CreatedAt:       transaction.CreatedAt,
			User: &types.UserBasicInfo{
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
			},
		}
	}

	response := &types.AdminTransactionResponse{
		Transactions: transactionsWithUsers,
		Pagination: types.PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}

	return response, nil
}

// GetTransactionDetailsAdminService returns detailed transaction information
func GetTransactionDetailsAdminService(transactionID string) (*types.TransactionDetailsResponse, error) {
	var transaction models.Transaction
	err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Get user details
	var user models.User
	database.DB.Where("user_id = ?", transaction.UserID).First(&user)

	// Get wallet details
	var wallet models.Wallet
	database.DB.Where("user_id = ?", transaction.UserID).First(&wallet)

	// var conversion *models.Conversions
	// // Get related conversion if it's a conversion transaction
	// if transaction.TransactionType == models.Conversion {
	// 	var conv models.Conversions
	// 	err := database.DB.Where("transaction_id = ?", transactionID).First(&conv).Error
	// 	if err == nil {
	// 		conversion = &conv
	// 	}
	// }

	// response := &types.TransactionDetailsResponse{
	// 	Transaction: transaction,
	// }

	return &types.TransactionDetailsResponse{
		Transaction: transaction,
	}, nil
}

// ApproveTransactionService approves a pending transaction
func ApproveTransactionService(transactionID string, adminID uint32) error {
	var transaction models.Transaction
	err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("transaction not found")
		}
		return err
	}

	if transaction.Status != models.TransactionPending {
		return errors.New("transaction is not pending")
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Update transaction status
		if err := tx.Model(&transaction).Update("status", models.TransactionCompleted).Error; err != nil {
			return err
		}

		// Update wallet balance if it's a deposit
		if transaction.TransactionType == models.Deposit {
			var wallet models.Wallet
			if err := tx.Where("user_id = ?", transaction.UserID).First(&wallet).Error; err != nil {
				return err
			}

			updates := map[string]interface{}{
				"total_deposits":      gorm.Expr("total_deposits + ?", transaction.TransactionDetails.FromAmount),
				"last_transaction_at": time.Now(),
			}

			if err := tx.Model(&wallet).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Log admin action
		adminLog := models.AdminLog{
			AdminID:  adminID,
			Action:   "approve_transaction",
			Target:   "transaction",
			TargetID: transactionID,
			Details: fmt.Sprint(map[string]interface{}{
				"transaction_id":  transactionID,
				"previous_status": "pending",
				"new_status":      "completed",
			}),
		}

		return tx.Create(&adminLog).Error
	})
}

// RejectTransactionService rejects a pending transaction
func RejectTransactionService(transactionID string, adminID uint32, reason string) error {
	var transaction models.Transaction
	err := database.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("transaction not found")
		}
		return err
	}

	if transaction.Status != models.TransactionPending {
		return errors.New("transaction is not pending")
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Update transaction status and description
		updates := map[string]interface{}{
			"status": models.TransactionFailed,
		}

		if reason != "" {
			updates["description"] = transaction.Description + " | Rejection reason: " + reason
		}

		if err := tx.Model(&transaction).Updates(updates).Error; err != nil {
			return err
		}

		// Refund wallet balance if needed (for withdrawal or conversion)
		if transaction.TransactionType == models.Withdrawal || transaction.TransactionType == models.Conversion {
			var wallet models.Wallet
			if err := tx.Where("user_id = ?", transaction.UserID).First(&wallet).Error; err != nil {
				return err
			}

			updates := map[string]interface{}{
				"last_transaction_at": time.Now(),
			}

			switch transaction.TransactionDetails.FromCurrency {
			case "NGN":
				updates["balance_ngn"] = gorm.Expr("balance_ngn + ?", transaction.TransactionDetails.FromAmount)
			case "GHS":
				updates["balance_ghs"] = gorm.Expr("balance_ghs + ?", transaction.TransactionDetails.FromAmount)
			}

			if err := tx.Model(&wallet).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Log admin action
		adminLog := models.AdminLog{
			AdminID:  adminID,
			Action:   "reject_transaction",
			Target:   "transaction",
			TargetID: transactionID,
			Details: fmt.Sprint(map[string]interface{}{
				"transaction_id":  transactionID,
				"reason":          reason,
				"previous_status": "pending",
				"new_status":      "failed",
			}),
		}

		return tx.Create(&adminLog).Error
	})
}

// GetAllUsersAdminService returns paginated users for admin
func GetAllUsersAdminService(query types.UserQueryParams) (*types.UsersResponse, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}

	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build filter conditions
	tx := database.DB.Model(&models.User{})

	switch query.Status {
	case "active":
		tx = tx.Where("is_verified = ?", true)
	case "inactive":
		tx = tx.Where("is_verified = ?", false)
	}

	switch query.Blocked {
	case "true":
		tx = tx.Where("is_blocked = ?", true)
	case "false":
		tx = tx.Where("is_blocked = ?", false)
	}

	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		tx = tx.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ? OR username ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	var total int64
	tx.Count(&total)

	// Get users (exclude password)
	var users []models.User
	tx.Select("id, created_at, updated_at, deleted_at, first_name, last_name, email, username, profile_picture, phone_number, is_admin, is_verified, is_blocked, user_id, country, is_two_factor_enabled").
		Order("created_at desc").Offset(offset).Limit(limit).Find(&users)

	// Get wallet details and transaction count for each user
	usersWithWallets := make([]types.UserWithWallet, len(users))
	for i, user := range users {
		var wallet models.Wallet
		database.DB.Where("user_id = ?", strconv.Itoa(int(user.UserID))).First(&wallet)

		var txnCount int64
		database.DB.Model(&models.Transaction{}).Where("user_id = ?", strconv.Itoa(int(user.UserID))).Count(&txnCount)

		walletBalance := &types.WalletBalance{
			Currency:         wallet.Currency,
			TotalDeposits:    wallet.TotalDeposits,
			TotalWithdrawals: wallet.TotalWithdrawals,
			TotalConversions: wallet.TotalConversions,
		}

		usersWithWallets[i] = types.UserWithWallet{
			User:             user,
			Wallet:           walletBalance,
			TransactionCount: txnCount,
		}
	}

	response := &types.UsersResponse{
		Users: usersWithWallets,
		Pagination: types.PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}

	return response, nil
}

// BlockUserService blocks a user account
func BlockUserService(userID string, adminID uint32, reason string) error {
	var user models.User
	err := database.DB.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if user.IsBlocked {
		return errors.New("user is already blocked")
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Block user
		updates := map[string]interface{}{
			"is_blocked":  true,
			"is_verified": false,
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return err
		}

		// Log admin action
		adminLog := models.AdminLog{
			AdminID:  adminID,
			Action:   "block_user",
			Target:   "user",
			TargetID: userID,
			Details: fmt.Sprint(map[string]interface{}{
				"user_id": userID,
				"reason":  reason,
			}),
		}

		return tx.Create(&adminLog).Error
	})
}

// UnblockUserService unblocks a user account
func UnblockUserService(userID string, adminID uint32) error {
	var user models.User
	err := database.DB.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if !user.IsBlocked {
		return errors.New("user is not blocked")
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Unblock user
		updates := map[string]interface{}{
			"is_blocked":  false,
			"is_verified": true,
		}

		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return err
		}

		// Log admin action
		adminLog := models.AdminLog{
			AdminID:  adminID,
			Action:   "unblock_user",
			Target:   "user",
			TargetID: userID,
			Details: fmt.Sprint(map[string]interface{}{
				"user_id": userID,
			}),
		}

		return tx.Create(&adminLog).Error
	})
}
