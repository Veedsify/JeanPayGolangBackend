package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
)

// GetDashboardOverview retrieves dashboard overview for a user
func GetDashboardOverview(userID uint) (*types.DashboardOverview, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Get wallet information
	wallets, err := GetWalletBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Get recent transactions (last 5)
	var recentTxns []models.Transaction
	if err := database.DB.Preload("TransactionDetails").Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentTxns).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	// Convert to RecentTxn format
	var recentTxnList []types.RecentTxn
	for _, txn := range recentTxns {
		recentTxnList = append(recentTxnList, types.RecentTxn{
			ID:            txn.TransactionID,
			Type:          string(txn.TransactionType),
			Status:        string(txn.Status),
			ToAmount:      txn.TransactionDetails.ToAmount,
			FromAmount:    txn.TransactionDetails.FromAmount,
			FromCurrency:  txn.TransactionDetails.FromCurrency,
			ToCurrency:    txn.TransactionDetails.ToCurrency,
			Description:   txn.Description,
			Recipient:     txn.TransactionDetails.RecipientName,
			TransactionID: txn.TransactionID,
			CreatedAt:     txn.CreatedAt,
		})
	}

	// Get transaction counts
	var totalTxns, pendingTxns, completedTxns int64
	database.DB.Model(&models.Transaction{}).Where("user_id = ?", userID).Count(&totalTxns)
	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&pendingTxns)
	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedTxns)

	// Get exchange rates
	ngnToGhs, _ := getCurrentExchangeRate("NGN", "GHS")
	ghsToNgn, _ := getCurrentExchangeRate("GHS", "NGN")

	// Set default rates if not found
	if ngnToGhs == 0 {
		ngnToGhs = 0.0053
	}
	if ghsToNgn == 0 {
		ghsToNgn = 188.68
	}

	// Calculate total balance in NGN equivalent
	var ngnBalance, ghsBalance float64
	var primaryCurrency string = "NGN"

	for _, wallet := range wallets {
		switch wallet.Currency {
		case "NGN":
			ngnBalance = wallet.Balance
		case "GHS":
			ghsBalance = wallet.Balance
		}
	}

	totalBalance := ngnBalance + (ghsBalance * ghsToNgn)

	overview := &types.DashboardOverview{
		Wallet: types.WalletSummary{
			Balance: ngnBalance, Currency: primaryCurrency,
			TotalBalance: libs.RoundCurrency(totalBalance),
		},
		RecentTxns: recentTxnList,
		ExchangeRates: types.ExchangeRateData{
			NGNToGHS: ngnToGhs,
			GHSToNGN: ghsToNgn,
		},
		QuickStats: types.QuickStatsData{
			TotalTransactions:   totalTxns,
			PendingTransactions: pendingTxns,
			CompletedTxns:       completedTxns,
		},
	}

	return overview, nil
}

// GetDashboardStats retrieves detailed dashboard statistics for a user
func GetDashboardStats(userID uint32) (*types.DashboardStats, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)
	endOfLastMonth := startOfMonth.Add(-time.Second)

	// Get monthly transaction counts
	var monthlyDeposits, monthlyWithdrawals, monthlyConversions int64

	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND transaction_type = ? AND created_at >= ?",
		userID, "deposit", startOfMonth).Count(&monthlyDeposits)

	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND transaction_type = ? AND created_at >= ?",
		userID, "withdrawal", startOfMonth).Count(&monthlyWithdrawals)

	database.DB.Model(&models.Transaction{}).Where("user_id = ? AND transaction_type = ? AND created_at >= ?",
		userID, "conversion", startOfMonth).Count(&monthlyConversions)

	// Get transaction volume for this month and last month
	var thisMonthResult, lastMonthResult struct {
		Total float64
	}

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0) as total").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.status = ? AND transactions.created_at >= ?", userID, "completed", startOfMonth).
		Scan(&thisMonthResult)

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0) as total").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.status = ? AND transactions.created_at >= ? AND transactions.created_at <= ?",
			userID, "completed", startOfLastMonth, endOfLastMonth).
		Scan(&lastMonthResult)

	thisMonth := thisMonthResult.Total
	lastMonth := lastMonthResult.Total

	// Calculate percentage change
	var percentageChange float64
	if lastMonth > 0 {
		percentageChange = ((thisMonth - lastMonth) / lastMonth) * 100
	}

	// Get conversion statistics
	var ngnToGhsConversions, ghsToNgnConversions int64

	database.DB.Model(&models.Conversions{}).Where("user_id = ? AND from_currency = ? AND to_currency = ? AND created_at >= ?",
		userID, "NGN", "GHS", startOfMonth).Count(&ngnToGhsConversions)

	database.DB.Model(&models.Conversions{}).Where("user_id = ? AND from_currency = ? AND to_currency = ? AND created_at >= ?",
		userID, "GHS", "NGN", startOfMonth).Count(&ghsToNgnConversions)

	// Get chart data (last 7 days)
	dailyTxnData := getDailyTransactionData(userID, 7)
	monthlyVolData := getMonthlyVolumeData(userID, 6)

	stats := &types.DashboardStats{
		MonthlyStats: types.MonthlyStatsData{
			Deposits:    monthlyDeposits,
			Withdrawals: monthlyWithdrawals,
			Conversions: monthlyConversions,
		},
		TransactionVol: types.TransactionVolData{
			ThisMonth:     libs.RoundCurrency(thisMonth),
			LastMonth:     libs.RoundCurrency(lastMonth),
			PercentChange: libs.RoundCurrency(percentageChange),
		},
		ConversionStats: types.ConversionStatsData{
			NGNToGHS: ngnToGhsConversions,
			GHSToNGN: ghsToNgnConversions,
		},
		ChartData: types.ChartData{
			DailyTransactions: dailyTxnData,
			MonthlyVolume:     monthlyVolData,
		},
	}

	return stats, nil
}

// getDailyTransactionData gets daily transaction data for chart
func getDailyTransactionData(userID uint32, days int) []types.DailyTxnData {
	var result []types.DailyTxnData

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Second)

		var count int64
		var volumeResult struct {
			Total float64
		}

		database.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startOfDay, endOfDay).
			Count(&count)

		database.DB.Table("transactions").
			Select("COALESCE(SUM(transaction_details.from_amount), 0) as total").
			Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
			Where("transactions.user_id = ? AND transactions.status = ? AND transactions.created_at >= ? AND transactions.created_at <= ?",
				userID, "completed", startOfDay, endOfDay).
			Scan(&volumeResult)

		result = append(result, types.DailyTxnData{
			Date:   date.Format("2006-01-02"),
			Count:  count,
			Volume: libs.RoundCurrency(volumeResult.Total),
		})
	}

	return result
}

// getMonthlyVolumeData gets monthly volume data for chart
func getMonthlyVolumeData(userID uint32, months int) []types.MonthlyVolData {
	var result []types.MonthlyVolData

	for i := months - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		var volumeResult struct {
			Total float64
		}

		database.DB.Table("transactions").
			Select("COALESCE(SUM(transaction_details.from_amount), 0) as total").
			Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
			Where("transactions.user_id = ? AND transactions.status = ? AND transactions.created_at >= ? AND transactions.created_at <= ?",
				userID, "completed", startOfMonth, endOfMonth).
			Scan(&volumeResult)

		result = append(result, types.MonthlyVolData{
			Month:  date.Format("2006-01"),
			Volume: libs.RoundCurrency(volumeResult.Total),
		})
	}

	return result
}

// GetUserTransactionSummary gets transaction summary for a user
func GetUserTransactionSummary(userID uint32) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var totalDeposits, totalWithdrawals, totalConversions float64
	var depositCount, withdrawalCount, conversionCount int64

	// Get deposit summary
	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ?", userID, "deposit", "completed").
		Scan(&totalDeposits)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, "deposit", "completed").
		Count(&depositCount)

	// Get withdrawal summary
	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ?", userID, "withdrawal", "completed").
		Scan(&totalWithdrawals)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, "withdrawal", "completed").
		Count(&withdrawalCount)

	// Get conversion summary
	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ?", userID, "conversion", "completed").
		Scan(&totalConversions)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND status = ?", userID, "conversion", "completed").
		Count(&conversionCount)

	summary := map[string]any{
		"deposits": map[string]any{
			"total": libs.RoundCurrency(totalDeposits),
			"count": depositCount,
		},
		"withdrawals": map[string]any{
			"total": libs.RoundCurrency(totalWithdrawals),
			"count": withdrawalCount,
		},
		"conversions": map[string]any{
			"total": libs.RoundCurrency(totalConversions),
			"count": conversionCount,
		},
	}

	return summary, nil
}

// GetDashboardSummary gets dashboard summary for a user
func GetDashboardSummary(userID uint) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Get wallet balances
	wallets, err := GetWalletBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Get total transactions this month
	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())

	var monthlyTransactions int64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfMonth).
		Count(&monthlyTransactions)

	// Get pending transactions
	var pendingTransactions int64
	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&pendingTransactions)

	// Get unread notifications
	var unreadNotifications int64
	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&unreadNotifications)

	// Calculate combined balance info
	var ngnBalance, ghsBalance float64
	for _, wallet := range wallets {
		switch wallet.Currency {
		case "NGN":
			ngnBalance = wallet.Balance
		case "GHS":
			ghsBalance = wallet.Balance
		}
	}

	summary := map[string]any{
		"ngn_balance":          libs.RoundCurrency(ngnBalance),
		"ghs_balance":          libs.RoundCurrency(ghsBalance),
		"monthly_transactions": monthlyTransactions,
		"pending_transactions": pendingTransactions,
		"unread_notifications": unreadNotifications,
		"primary_currency":     "NGN",
	}

	return summary, nil
}

// GetRecentActivity gets recent activity for a user
func GetRecentActivity(userID uint32) ([]map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Get recent transactions
	var transactions []models.Transaction
	err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&transactions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent transactions: %w", err)
	}

	var activities []map[string]interface{}
	for _, txn := range transactions {
		activity := map[string]any{
			"id":          txn.TransactionID,
			"type":        string(txn.TransactionType),
			"status":      string(txn.Status),
			"description": txn.Description,
			"created_at":  txn.CreatedAt,
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// GetWalletOverview gets wallet overview for a user
func GetWalletOverview(userID uint) (map[string]any, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	wallets, err := GetWalletBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Get exchange rates
	totalBalance, err := GetUserTotalBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user total balance: %w", err)
	}
	ngnToGhs, _ := getCurrentExchangeRate("NGN", "GHS")
	ghsToNgn, _ := getCurrentExchangeRate("GHS", "NGN")

	if ngnToGhs == 0 {
		ngnToGhs = 0.0053
	}
	if ghsToNgn == 0 {
		ghsToNgn = 188.68
	}

	// Create wallet overview with both currencies
	var ngnWallet, ghsWallet *types.WalletBalance
	for i, wallet := range wallets {
		switch wallet.Currency {
		case "NGN":
			ngnWallet = &wallets[i]
		case "GHS":
			ghsWallet = &wallets[i]
		}
	}

	overview := map[string]any{
		"exchange_rates": map[string]float64{
			"NGN-GHS": ngnToGhs,
			"GHS-NGN": ghsToNgn,
		},
	}

	// Add total Balance info
	overview["totalBalance"] = libs.RoundCurrency(totalBalance)

	// Add NGN wallet info if exists
	if ngnWallet != nil {
		overview["ngn_wallet"] = map[string]any{
			"balance":           libs.RoundCurrency(ngnWallet.Balance),
			"total_deposits":    libs.RoundCurrency(ngnWallet.TotalDeposits),
			"total_withdrawals": libs.RoundCurrency(ngnWallet.TotalWithdrawals),
			"total_conversions": libs.RoundCurrency(ngnWallet.TotalConversions),
			"is_active":         ngnWallet.IsActive,
		}
	} else {
		overview["ngn_wallet"] = map[string]any{
			"balance":           0.0,
			"total_deposits":    0.0,
			"total_withdrawals": 0.0,
			"total_conversions": 0.0,
			"is_active":         true,
		}
	}

	// Add GHS wallet info if exists
	if ghsWallet != nil {
		overview["ghs_wallet"] = map[string]any{
			"balance":           libs.RoundCurrency(ghsWallet.Balance),
			"total_deposits":    libs.RoundCurrency(ghsWallet.TotalDeposits),
			"total_withdrawals": libs.RoundCurrency(ghsWallet.TotalWithdrawals),
			"total_conversions": libs.RoundCurrency(ghsWallet.TotalConversions),
			"is_active":         ghsWallet.IsActive,
		}
	} else {
		overview["ghs_wallet"] = map[string]any{
			"balance":           0.0,
			"total_deposits":    0.0,
			"total_withdrawals": 0.0,
			"total_conversions": 0.0,
			"is_active":         true,
		}
	}

	return overview, nil
}

// GetConversionStats gets conversion statistics for a user
func GetConversionStats(userID uint32) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())

	// Get conversion counts
	var totalConversions, monthlyConversions int64
	var ngnToGhsCount, ghsToNgnCount int64

	database.DB.Model(&models.Conversions{}).
		Where("user_id = ?", userID).
		Count(&totalConversions)

	database.DB.Model(&models.Conversions{}).
		Where("user_id = ? AND created_at >= ?", userID, startOfMonth).
		Count(&monthlyConversions)

	database.DB.Model(&models.Conversions{}).
		Where("user_id = ? AND from_currency = ? AND to_currency = ?", userID, "NGN", "GHS").
		Count(&ngnToGhsCount)

	database.DB.Model(&models.Conversions{}).
		Where("user_id = ? AND from_currency = ? AND to_currency = ?", userID, "GHS", "NGN").
		Count(&ghsToNgnCount)

	// Get conversion amounts
	var totalAmount, monthlyAmount float64

	database.DB.Model(&models.Conversions{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = ?", userID, "completed").
		Scan(&totalAmount)

	database.DB.Model(&models.Conversions{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, "completed", startOfMonth).
		Scan(&monthlyAmount)

	stats := map[string]any{
		"total_conversions":   totalConversions,
		"monthly_conversions": monthlyConversions,
		"ngn_to_ghs_count":    ngnToGhsCount,
		"ghs_to_ngn_count":    ghsToNgnCount,
		"total_amount":        libs.RoundCurrency(totalAmount),
		"monthly_amount":      libs.RoundCurrency(monthlyAmount),
	}

	return stats, nil
}

// GetMonthlyStats gets monthly statistics for a user
func GetMonthlyStats(userID uint32) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	currentDate := time.Now()
	startOfMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())

	// Get monthly transaction counts by type
	var deposits, withdrawals, conversions int64
	var depositAmount, withdrawalAmount, conversionAmount float64

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND created_at >= ?", userID, "deposit", startOfMonth).
		Count(&deposits)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND created_at >= ?", userID, "withdrawal", startOfMonth).
		Count(&withdrawals)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND created_at >= ?", userID, "conversion", startOfMonth).
		Count(&conversions)

	// Get amounts
	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ? AND transactions.created_at >= ?", userID, "deposit", "completed", startOfMonth).
		Scan(&depositAmount)

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ? AND transactions.created_at >= ?", userID, "withdrawal", "completed", startOfMonth).
		Scan(&withdrawalAmount)

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ? AND transactions.created_at >= ?", userID, "conversion", "completed", startOfMonth).
		Scan(&conversionAmount)

	stats := map[string]any{
		"deposits": map[string]any{
			"count":  deposits,
			"amount": libs.RoundCurrency(depositAmount),
		},
		"withdrawals": map[string]any{
			"count":  withdrawals,
			"amount": libs.RoundCurrency(withdrawalAmount),
		},
		"conversions": map[string]any{
			"count":  conversions,
			"amount": libs.RoundCurrency(conversionAmount),
		},
		"month": currentDate.Format("2006-01"),
	}

	return stats, nil
}

// GetTransactionTrends gets transaction trends for a user
func GetTransactionTrends(userID uint32) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Get last 12 months of data
	var monthlyData []map[string]interface{}

	for i := 11; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		var count int64
		var amount float64

		database.DB.Model(&models.Transaction{}).
			Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startOfMonth, endOfMonth).
			Count(&count)

		database.DB.Table("transactions").
			Select("COALESCE(SUM(transaction_details.from_amount), 0)").
			Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
			Where("transactions.user_id = ? AND transactions.status = ? AND transactions.created_at >= ? AND transactions.created_at <= ?", userID, "completed", startOfMonth, endOfMonth).
			Scan(&amount)

		monthData := map[string]any{
			"month":  date.Format("2006-01"),
			"count":  count,
			"amount": libs.RoundCurrency(amount),
		}

		monthlyData = append(monthlyData, monthData)
	}

	trends := map[string]any{
		"monthly_data": monthlyData,
		"period":       "12_months",
	}

	return trends, nil
}

// GetDashboardTransactionStats retrieves transaction statistics for a specific period
func GetDashboardTransactionStats(userID uint32, period string) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var startDate time.Time
	now := time.Now()

	// Determine date range based on period
	switch period {
	case "day":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	// Get transaction counts by status
	var completedCount, pendingCount, failedCount int64

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, "completed", startDate).
		Count(&completedCount)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, "pending", startDate).
		Count(&pendingCount)

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, "failed", startDate).
		Count(&failedCount)

	// Get transaction amounts by type
	var totalIncome, totalExpense float64

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type IN (?, ?) AND transactions.status = ? AND transactions.created_at >= ?",
			userID, "deposit", "transfer", "completed", startDate).
		Scan(&totalIncome)

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type IN (?, ?) AND transactions.status = ? AND transactions.created_at >= ?",
			userID, "withdrawal", "conversion", "completed", startDate).
		Scan(&totalExpense)

	// Get conversion statistics
	var conversionCount int64
	var conversionAmount float64

	database.DB.Model(&models.Transaction{}).
		Where("user_id = ? AND transaction_type = ? AND status = ? AND created_at >= ?",
			userID, "conversion", "completed", startDate).
		Count(&conversionCount)

	database.DB.Table("transactions").
		Select("COALESCE(SUM(transaction_details.from_amount), 0)").
		Joins("LEFT JOIN transaction_details ON transactions.id = transaction_details.transaction_id").
		Where("transactions.user_id = ? AND transactions.transaction_type = ? AND transactions.status = ? AND transactions.created_at >= ?",
			userID, "conversion", "completed", startDate).
		Scan(&conversionAmount)

	stats := map[string]any{
		"period": period,
		"date_range": map[string]any{
			"start": startDate.Format("2006-01-02"),
			"end":   now.Format("2006-01-02"),
		},
		"transaction_counts": map[string]any{
			"completed": completedCount,
			"pending":   pendingCount,
			"failed":    failedCount,
			"total":     completedCount + pendingCount + failedCount,
		},
		"amounts": map[string]any{
			"income":  libs.RoundCurrency(totalIncome),
			"expense": libs.RoundCurrency(totalExpense),
			"net":     libs.RoundCurrency(totalIncome - totalExpense),
		},
		"conversions": map[string]any{
			"count":  conversionCount,
			"amount": libs.RoundCurrency(conversionAmount),
		},
	}

	return stats, nil
}
