package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// GetAllWithdrawals retrieves all withdrawal requests with pagination and filters
func GetAllWithdrawals(params types.WithdrawalQueryParams) ([]types.WithdrawalWithUser, types.WithdrawalPagination, error) {
	var withdrawals []types.WithdrawalWithUser
	var total int64

	// Build query for withdrawal transactions
	query := database.DB.Model(&models.Transaction{}).
		Select(`
			transactions.id,
			transactions.transaction_id,
			transactions.user_id,
			COALESCE(transaction_details.from_amount, 0) as amount,
			COALESCE(transaction_details.from_currency, '') as currency,
			0 as fee,
			COALESCE(transaction_details.from_amount, 0) as net_amount,
			transactions.status,
			transactions.payment_type as method,
			users.first_name,
			users.last_name,
			users.email,
			users.phone_number,
			users.profile_picture,
			CASE WHEN users.is_verified THEN 'verified' ELSE 'pending' END as verification_status,
			transaction_details.account_number,
			transaction_details.bank_name,
			transaction_details.phone_number as detail_phone,
			transaction_details.network,
			transaction_details.recipient_name
		`).
		Joins("LEFT JOIN users ON transactions.user_id = users.id").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.transaction_type = ? AND transactions.direction = ?", "WITHDRAWAL", "WITHDRAWAL")

	// Apply filters
	if params.Status != "" {
		query = query.Where("transactions.status = ?", strings.ToUpper(params.Status))
	}
	if params.Currency != "" {
		query = query.Where("transaction_details.from_currency = ?", strings.ToUpper(params.Currency))
	}
	if params.Method != "" {
		if strings.ToUpper(params.Method) == "MOMO" {
			query = query.Where("transactions.payment_type LIKE ?", "%MOMO%")
		} else if strings.ToUpper(params.Method) == "BANK" {
			query = query.Where("transactions.payment_type NOT LIKE ?", "%MOMO%")
		}
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where(`
			transactions.transaction_id LIKE ? OR 
			users.first_name LIKE ? OR 
			users.last_name LIKE ? OR 
			users.email LIKE ? OR
			CONCAT(users.first_name, ' ', users.last_name) LIKE ?
		`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}
	if params.FromDate != "" {
		query = query.Where("DATE(transactions.created_at) >= ?", params.FromDate)
	}
	if params.ToDate != "" {
		query = query.Where("DATE(transactions.created_at) <= ?", params.ToDate)
	}
	if params.MinAmount != nil {
		query = query.Where("transaction_details.from_amount >= ?", *params.MinAmount)
	}
	if params.MaxAmount != nil {
		query = query.Where("transaction_details.from_amount <= ?", *params.MaxAmount)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, types.WithdrawalPagination{}, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit).Order("transactions.created_at DESC")

	// Execute query
	rows, err := query.Rows()
	if err != nil {
		return nil, types.WithdrawalPagination{}, err
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var w types.WithdrawalWithUser
		var accountNumber, bankName, detailPhone, network, recipientName *string

		err := rows.Scan(
			&w.ID, &w.TransactionID, &w.UserID, &w.Amount, &w.Currency,
			&w.Fee, &w.NetAmount, &w.Status, &w.Method, &w.ReceiptURL,
			&w.RejectionReason, &w.CreatedAt, &w.UpdatedAt,
			&w.User.FirstName, &w.User.LastName, &w.User.Email,
			&w.User.PhoneNumber, &w.User.CountryCode, &w.User.ProfilePicture,
			&w.User.VerificationStatus, &accountNumber, &bankName,
			&detailPhone, &network, &recipientName,
		)
		if err != nil {
			return nil, types.WithdrawalPagination{}, err
		}

		// Set user ID
		w.User.UserID = w.UserID

		// Map method based on payment type
		if strings.Contains(strings.ToLower(w.Method), "momo") {
			w.Method = "MOMO"
		} else {
			w.Method = "BANK"
		}

		// Set account details
		w.AccountDetails = types.WithdrawalAccountDetails{}
		if accountNumber != nil {
			w.AccountDetails.AccountNumber = *accountNumber
		}
		if bankName != nil {
			w.AccountDetails.BankName = *bankName
		}
		if detailPhone != nil {
			w.AccountDetails.PhoneNumber = *detailPhone
		}
		if network != nil {
			w.AccountDetails.Network = *network
		}
		if recipientName != nil {
			// Note: RecipientName field should be added to WithdrawalAccountDetails type
			// For now, we'll use AccountName as a fallback
			w.AccountDetails.AccountName = *recipientName
		}

		withdrawals = append(withdrawals, w)
	}

	// Calculate pagination
	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))
	pagination := types.WithdrawalPagination{
		Page:  params.Page,
		Limit: params.Limit,
		Total: int(total),
		Pages: totalPages,
	}

	return withdrawals, pagination, nil
}

// GetWithdrawalStats retrieves withdrawal statistics
func GetWithdrawalStats() (types.WithdrawalStats, error) {
	var stats types.WithdrawalStats

	// Base query for withdrawals
	baseQuery := database.DB.Model(&models.Transaction{}).
		Where("transaction_type = ? AND direction = ?", "WITHDRAWAL", "WITHDRAWAL")

	// Total requests
	if err := baseQuery.Count(&stats.TotalRequests).Error; err != nil {
		return stats, err
	}

	// Status counts
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := baseQuery.
		Select("status", "COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return stats, err
	}

	for _, sc := range statusCounts {
		switch strings.ToUpper(sc.Status) {
		case "PENDING":
			stats.PendingRequests = sc.Count
		case "COMPLETED":
			stats.CompletedRequests = sc.Count
		case "FAILED":
			stats.FailedRequests = sc.Count
		}
	}

	// Total amount (join with transaction_details)
	var totalAmount float64
	if err := baseQuery.
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Scan(&totalAmount).Error; err != nil {
		return stats, err
	}
	stats.TotalAmount = totalAmount

	// Helper to get period stats
	getPeriodStats := func(start time.Time) (types.WithdrawalPeriodStats, error) {
		var count int64
		var amount float64

		query := database.DB.Model(&models.Transaction{}).
			Where("transaction_type = ? AND direction = ? AND created_at >= ?", "WITHDRAWAL", "WITHDRAWAL", start).
			Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id")

		if err := query.
			Count(&count).
			Select("COALESCE(SUM(transaction_details.from_amount), 0)").
			Scan(&amount).Error; err != nil {
			return types.WithdrawalPeriodStats{}, err
		}

		return types.WithdrawalPeriodStats{
			Count:  count,
			Amount: amount,
		}, nil
	}

	now := time.Now().Truncate(24 * time.Hour)

	// Today
	todayStart := now
	stats.Today, _ = getPeriodStats(todayStart)

	// This week (Monday start)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	stats.ThisWeek, _ = getPeriodStats(weekStart)

	// This month
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	stats.ThisMonth, _ = getPeriodStats(monthStart)

	return stats, nil
}

// GetAdminWithdrawalDetails retrieves detailed information for a specific withdrawal
func GetAdminWithdrawalDetails(withdrawalId string) (types.WithdrawalDetailsResponse, error) {
	var withdrawal types.WithdrawalDetailsResponse

	// Get withdrawal transaction with user and details
	query := database.DB.Model(&models.Transaction{}).
		Select(`
			transactions.*,
			users.first_name,
			users.last_name,
			users.email,
			users.phone_number,
			users.profile_picture,
			CASE WHEN users.is_verified THEN 'verified' ELSE 'pending' END as verification_status,
			transaction_details.account_number,
			transaction_details.bank_name,
			transaction_details.phone_number as detail_phone,
			transaction_details.network,
			transaction_details.recipient_name,
			transaction_details.from_amount,
			transaction_details.from_currency
		`).
		Joins("LEFT JOIN users ON transactions.user_id = users.id").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.transaction_id = ? AND transactions.transaction_type = ? AND transactions.direction = ?",
			withdrawalId, "WITHDRAWAL", "WITHDRAWAL")

	var transaction models.Transaction
	var user types.WithdrawalUser
	var accountNumber, bankName, detailPhone, network, recipientName *string
	var amount float64
	var currency string

	row := query.Row()
	err := row.Scan(
		&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.DeletedAt,
		&transaction.Code, &transaction.UserID, &transaction.TransactionID, &transaction.PaymentType,
		&transaction.Reason, &transaction.Status, &transaction.TransactionType, &transaction.Reference,
		&transaction.Direction, &transaction.Description, &transaction.ReceiptURL, &transaction.WithdrawalFee,
		&transaction.FeeCalculation, &transaction.CurrentRate,
		&user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber,
		&user.CountryCode, &user.ProfilePicture, &user.VerificationStatus,
		&accountNumber, &bankName, &detailPhone, &network, &recipientName,
		&amount, &currency,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return withdrawal, errors.New("withdrawal not found")
		}
		return withdrawal, err
	}

	// Map to withdrawal response
	withdrawal.ID = int(transaction.ID)
	withdrawal.TransactionID = transaction.TransactionID
	withdrawal.UserID = int(transaction.UserID)
	withdrawal.Amount = amount
	withdrawal.Currency = currency
	withdrawal.Fee = transaction.WithdrawalFee
	withdrawal.NetAmount = withdrawal.Amount - withdrawal.Fee
	withdrawal.Status = string(transaction.Status)
	withdrawal.Method = string(transaction.PaymentType)
	withdrawal.ReceiptURL = transaction.ReceiptURL
	withdrawal.RejectionReason = transaction.Reason
	withdrawal.CreatedAt = transaction.CreatedAt.Format(time.RFC3339)
	withdrawal.UpdatedAt = transaction.UpdatedAt.Format(time.RFC3339)

	// Set user info
	withdrawal.User = user
	withdrawal.User.UserID = int(transaction.UserID)

	// Set account details
	withdrawal.AccountDetails = types.WithdrawalAccountDetails{}
	if accountNumber != nil {
		withdrawal.AccountDetails.AccountNumber = *accountNumber
	}
	if bankName != nil {
		withdrawal.AccountDetails.BankName = *bankName
	}
	if detailPhone != nil {
		withdrawal.AccountDetails.PhoneNumber = *detailPhone
	}
	if network != nil {
		withdrawal.AccountDetails.Network = *network
	}
	if recipientName != nil {
		// Note: RecipientName field should be added to WithdrawalAccountDetails type
		// For now, we'll use AccountName as a fallback
		withdrawal.AccountDetails.AccountName = *recipientName
	}

	return withdrawal, nil
}

// ApproveWithdrawal approves a pending withdrawal request
func ApproveWithdrawal(withdrawalId, adminNotes string, adminID uint) (types.WithdrawalActionResponse, error) {
	// Find the withdrawal transaction
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ? AND transaction_type = ? AND direction = ?",
		withdrawalId, "WITHDRAWAL", "WITHDRAWAL").First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.WithdrawalActionResponse{}, errors.New("withdrawal not found")
		}
		return types.WithdrawalActionResponse{}, err
	}

	// Check if withdrawal is in pending status
	if transaction.Status != "PENDING" {
		return types.WithdrawalActionResponse{}, errors.New("withdrawal is not in pending status")
	}

	// Update transaction status
	updates := map[string]interface{}{
		"status":     "COMPLETED",
		"updated_at": time.Now(),
	}

	if adminNotes != "" {
		updates["description"] = adminNotes
	}

	if err := database.DB.Model(&transaction).Updates(updates).Error; err != nil {
		return types.WithdrawalActionResponse{}, err
	}

	// Log admin action
	logAdminAction(adminID, "APPROVE_WITHDRAWAL", "withdrawal", withdrawalId,
		fmt.Sprintf("Approved withdrawal %s", withdrawalId))

	return types.WithdrawalActionResponse{
		Success:   true,
		Message:   "Withdrawal approved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// RejectWithdrawal rejects a pending withdrawal request
func RejectWithdrawal(withdrawalId, reason, adminNotes string, adminID uint) (types.WithdrawalActionResponse, error) {
	// Find the withdrawal transaction
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ? AND transaction_type = ? AND direction = ?",
		withdrawalId, "WITHDRAWAL", "WITHDRAWAL").First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.WithdrawalActionResponse{}, errors.New("withdrawal not found")
		}
		return types.WithdrawalActionResponse{}, err
	}

	// Check if withdrawal is in pending status
	if transaction.Status != "PENDING" {
		return types.WithdrawalActionResponse{}, errors.New("withdrawal is not in pending status")
	}

	// Update transaction status
	updates := map[string]interface{}{
		"status":     "FAILED",
		"reason":     reason,
		"updated_at": time.Now(),
	}

	if adminNotes != "" {
		updates["description"] = adminNotes
	}

	if err := database.DB.Model(&transaction).Updates(updates).Error; err != nil {
		return types.WithdrawalActionResponse{}, err
	}

	// Log admin action
	logAdminAction(adminID, "REJECT_WITHDRAWAL", "withdrawal", withdrawalId,
		fmt.Sprintf("Rejected withdrawal %s: %s", withdrawalId, reason))

	return types.WithdrawalActionResponse{
		Success:   true,
		Message:   "Withdrawal rejected successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// UpdateWithdrawalStatus updates the status of a withdrawal request
func UpdateWithdrawalStatus(withdrawalId, status, reason, adminNotes string, adminID uint) (types.WithdrawalActionResponse, error) {
	// Find the withdrawal transaction
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ? AND transaction_type = ? AND direction = ?",
		withdrawalId, "WITHDRAWAL", "WITHDRAWAL").First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.WithdrawalActionResponse{}, errors.New("withdrawal not found")
		}
		return types.WithdrawalActionResponse{}, err
	}

	// Validate status
	validStatuses := []string{"PENDING", "COMPLETED", "FAILED", "CANCELLED"}
	isValid := false
	for _, validStatus := range validStatuses {
		if strings.ToUpper(status) == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return types.WithdrawalActionResponse{}, errors.New("invalid status")
	}

	// Update transaction
	updates := map[string]interface{}{
		"status":     strings.ToUpper(status),
		"updated_at": time.Now(),
	}

	if reason != "" {
		updates["reason"] = reason
	}
	if adminNotes != "" {
		updates["description"] = adminNotes
	}

	if err := database.DB.Model(&transaction).Updates(updates).Error; err != nil {
		return types.WithdrawalActionResponse{}, err
	}

	// Log admin action
	logAdminAction(adminID, "UPDATE_WITHDRAWAL_STATUS", "withdrawal", withdrawalId,
		fmt.Sprintf("Updated withdrawal %s status to %s", withdrawalId, status))

	return types.WithdrawalActionResponse{
		Success:   true,
		Message:   "Withdrawal status updated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// AddWithdrawalNote adds an admin note to a withdrawal request
func AddWithdrawalNote(withdrawalId, note string, adminID uint) (types.WithdrawalActionResponse, error) {
	// Find the withdrawal transaction
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ? AND transaction_type = ? AND direction = ?",
		withdrawalId, "WITHDRAWAL", "WITHDRAWAL").First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return types.WithdrawalActionResponse{}, errors.New("withdrawal not found")
		}
		return types.WithdrawalActionResponse{}, err
	}

	// For now, we'll append the note to the description field
	// In a production system, you might want a separate notes table
	currentDescription := transaction.Description
	newDescription := currentDescription
	if currentDescription != "" {
		newDescription = currentDescription + "\n\nAdmin Note: " + note
	} else {
		newDescription = "Admin Note: " + note
	}

	if err := database.DB.Model(&transaction).Update("description", newDescription).Error; err != nil {
		return types.WithdrawalActionResponse{}, err
	}

	// Log admin action
	logAdminAction(adminID, "ADD_WITHDRAWAL_NOTE", "withdrawal", withdrawalId,
		fmt.Sprintf("Added note to withdrawal %s", withdrawalId))

	return types.WithdrawalActionResponse{
		Success:   true,
		Message:   "Withdrawal note added successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// GetWithdrawalNotes retrieves all notes for a withdrawal request
func GetWithdrawalNotes(withdrawalId string) ([]types.WithdrawalNote, error) {
	// For now, return empty array since we don't have a separate notes table
	// In a production system, you would query a withdrawal_notes table
	return []types.WithdrawalNote{}, nil
}

// GetWithdrawalHistory retrieves status history for a withdrawal request
func GetWithdrawalHistory(withdrawalId string) ([]types.WithdrawalStatusHistory, error) {
	// For now, return empty array since we don't have a separate history table
	// In a production system, you would query a withdrawal_status_history table
	return []types.WithdrawalStatusHistory{}, nil
}

// Helper function to log admin actions
func logAdminAction(adminID uint, action, target, targetID, details string) {
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   action,
		Target:   target,
		TargetID: targetID,
		Details:  details,
	}
	database.DB.Create(&adminLog)
}
