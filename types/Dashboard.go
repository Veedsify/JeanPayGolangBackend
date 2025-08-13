package types

import "time"

// DashboardOverview represents dashboard overview data
type DashboardOverview struct {
	Wallet        WalletSummary    `json:"wallet"`
	RecentTxns    []RecentTxn      `json:"recentTransactions"`
	ExchangeRates ExchangeRateData `json:"exchangeRates"`
	QuickStats    QuickStatsData   `json:"quickStats"`
}

// DashboardStats represents detailed dashboard statistics
type DashboardStats struct {
	MonthlyStats    MonthlyStatsData    `json:"monthlyStats"`
	TransactionVol  TransactionVolData  `json:"transactionVolume"`
	ConversionStats ConversionStatsData `json:"conversionStats"`
	ChartData       ChartData           `json:"chartData"`
}

// WalletSummary represents wallet summary for dashboard
type WalletSummary struct {
	Balance      float64 `json:"balance"`
	Currency     string  `json:"currency"`
	TotalBalance float64 `json:"totalBalance"`
}

// RecentTxn represents recent transaction for dashboard
type RecentTxn struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	ToAmount      float64   `json:"toAmount"`
	FromAmount    float64   `json:"fromAmount"`
	FromCurrency  string    `json:"fromCurrency"`
	ToCurrency    string    `json:"toCurrency"`
	Recipient     string    `json:"recipient"`
	TransactionID string    `json:"transactionId"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"createdAt"`
}

// ExchangeRateData represents exchange rate information
type ExchangeRateData struct {
	NGNToGHS float64 `json:"ngnToGhs"`
	GHSToNGN float64 `json:"ghsToNgn"`
}

// QuickStatsData represents quick statistics
type QuickStatsData struct {
	TotalTransactions   int64 `json:"totalTransactions"`
	PendingTransactions int64 `json:"pendingTransactions"`
	CompletedTxns       int64 `json:"completedTransactions"`
}

// MonthlyStatsData represents monthly statistics
type MonthlyStatsData struct {
	Deposits    int64 `json:"deposits"`
	Withdrawals int64 `json:"withdrawals"`
	Conversions int64 `json:"conversions"`
}

// TransactionVolData represents transaction volume data
type TransactionVolData struct {
	ThisMonth     float64 `json:"thisMonth"`
	LastMonth     float64 `json:"lastMonth"`
	PercentChange float64 `json:"percentageChange"`
}

// ConversionStatsData represents conversion statistics
type ConversionStatsData struct {
	NGNToGHS int64 `json:"ngnToGhs"`
	GHSToNGN int64 `json:"ghsToNgn"`
}

// ChartData represents chart data for dashboard
type ChartData struct {
	DailyTransactions []DailyTxnData   `json:"dailyTransactions"`
	MonthlyVolume     []MonthlyVolData `json:"monthlyVolume"`
}

// DailyTxnData represents daily transaction data
type DailyTxnData struct {
	Date   string  `json:"date"`
	Count  int64   `json:"count"`
	Volume float64 `json:"volume"`
}

// MonthlyVolData represents monthly volume data
type MonthlyVolData struct {
	Month  string  `json:"month"`
	Volume float64 `json:"volume"`
}
