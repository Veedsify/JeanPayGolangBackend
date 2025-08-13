package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateWalletRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required"`
}

type UpdateWalletRequest struct {
	BalanceNGN       *float64 `json:"balance_ngn" form:"balance_ngn"`
	BalanceGHS       *float64 `json:"balance_ghs" form:"balance_ghs"`
	TotalDeposits    *float64 `json:"total_deposits" form:"total_deposits"`
	TotalWithdrawals *float64 `json:"total_withdrawals" form:"total_withdrawals"`
	TotalConversions *float64 `json:"total_conversions" form:"total_conversions"`
	IsActive         *bool    `json:"is_active" form:"is_active"`
}

type WalletResponse struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	Currency          string     `json:"balance_ngn" gorm:"enum('NGN', 'GHS');default('NGN')"`
	Balance           float64    `json:"balance" `
	TotalDeposits     float64    `json:"total_deposits"`
	TotalWithdrawals  float64    `json:"total_withdrawals"`
	TotalConversions  float64    `json:"total_conversions"`
	IsActive          bool       `json:"is_active"`
	LastTransactionAt *time.Time `json:"last_transaction_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type GetWalletsRequest struct {
	UserID   string `form:"user_id"`
	IsActive *bool  `form:"is_active"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

type GetWalletsResponse struct {
	Wallets    []WalletResponse `json:"wallets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type WalletBalanceRequest struct {
	UserID   string  `json:"user_id" form:"user_id" binding:"required"`
	Amount   float64 `json:"amount" form:"amount" binding:"required"`
	Currency string  `json:"currency" form:"currency" binding:"required"`
	Type     string  `json:"type" form:"type" binding:"required"` // "credit" or "debit"
}

type WalletBalanceResponse struct {
	UserID        string    `json:"user_id"`
	PreviousNGN   float64   `json:"previous_ngn"`
	PreviousGHS   float64   `json:"previous_ghs"`
	NewBalanceNGN float64   `json:"new_balance_ngn"`
	NewBalanceGHS float64   `json:"new_balance_ghs"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WalletStatsResponse struct {
	TotalUsers            int64   `json:"total_users"`
	TotalBalanceNGN       float64 `json:"total_balance_ngn"`
	TotalBalanceGHS       float64 `json:"total_balance_ghs"`
	TotalDeposits         float64 `json:"total_deposits"`
	TotalWithdrawals      float64 `json:"total_withdrawals"`
	TotalConversions      float64 `json:"total_conversions"`
	ActiveWallets         int64   `json:"active_wallets"`
	InactiveWallets       int64   `json:"inactive_wallets"`
	WalletsWithBalanceNGN int64   `json:"wallets_with_balance_ngn"`
	WalletsWithBalanceGHS int64   `json:"wallets_with_balance_ghs"`
}

type TransferRequest struct {
	FromUserID string  `json:"from_user_id" form:"from_user_id" binding:"required"`
	ToUserID   string  `json:"to_user_id" form:"to_user_id" binding:"required"`
	Amount     float64 `json:"amount" form:"amount" binding:"required"`
	Currency   string  `json:"currency" form:"currency" binding:"required"`
	Reference  string  `json:"reference" form:"reference" binding:"required"`
}

type TransferResponse struct {
	TransactionID  string    `json:"transaction_id"`
	FromUserID     string    `json:"from_user_id"`
	ToUserID       string    `json:"to_user_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Reference      string    `json:"reference"`
	FromBalanceNGN float64   `json:"from_balance_ngn"`
	FromBalanceGHS float64   `json:"from_balance_ghs"`
	ToBalanceNGN   float64   `json:"to_balance_ngn"`
	ToBalanceGHS   float64   `json:"to_balance_ghs"`
	TransferredAt  time.Time `json:"transferred_at"`
}

func ToWalletResponse(wallet *models.Wallet) WalletResponse {
	return WalletResponse{
		ID:                wallet.ID,
		UserID:            wallet.UserID,
		Currency:          wallet.Currency,
		TotalDeposits:     wallet.TotalDeposits,
		TotalWithdrawals:  wallet.TotalWithdrawals,
		TotalConversions:  wallet.TotalConversions,
		IsActive:          wallet.IsActive,
		LastTransactionAt: wallet.LastTransactionAt,
		CreatedAt:         wallet.CreatedAt,
		UpdatedAt:         wallet.UpdatedAt,
	}
}

func ToWalletsResponse(wallets []models.Wallet) []WalletResponse {
	var response []WalletResponse
	for _, wallet := range wallets {
		response = append(response, ToWalletResponse(&wallet))
	}
	return response
}

// WalletBalance represents wallet balance information
type WalletBalance struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"userId"`
	Balance           float64    `json:"balance"`
	Currency          string     `json:"currency"`
	WalletID          uint64     `json:"walletId"`
	TotalDeposits     float64    `json:"totalDeposits"`
	TotalWithdrawals  float64    `json:"totalWithdrawals"`
	TotalConversions  float64    `json:"totalConversions"`
	IsActive          bool       `json:"isActive"`
	LastTransactionAt *time.Time `json:"lastTransactionAt,omitempty"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// TopUpRequest represents a wallet top-up request
type TopUpRequest struct {
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	Currency         string  `json:"currency" validate:"required,oneof=NGN GHS"`
	PaymentMethod    string  `json:"paymentMethod" validate:"required,oneof=bank momo"`
	PaymentReference string  `json:"paymentReference,omitempty"`
	IsDirectPayment  bool    `json:"isDirectPayment,omitempty"`
}

// WithdrawRequest represents a wallet withdrawal request
type WithdrawRequest struct {
	Amount           float64                `json:"amount" validate:"required,gt=0"`
	Currency         string                 `json:"currency" validate:"required,oneof=NGN GHS"`
	WithdrawalMethod string                 `json:"withdrawalMethod" validate:"required,oneof=bank momo"`
	AccountDetails   map[string]interface{} `json:"accountDetails" validate:"required"`
}

// TopUpResponse represents the response for a top-up request
type TopUpResponse struct {
	TransactionID    string    `json:"transactionId"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	PaymentMethod    string    `json:"paymentMethod"`
	Status           string    `json:"status"`
	PaymentReference string    `json:"paymentReference"`
	CreatedAt        time.Time `json:"createdAt"`
}

// WithdrawResponse represents the response for a withdrawal request
type WithdrawResponse struct {
	TransactionID    string                 `json:"transactionId"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	WithdrawalMethod string                 `json:"withdrawalMethod"`
	Status           string                 `json:"status"`
	AccountDetails   map[string]interface{} `json:"accountDetails"`
	CreatedAt        time.Time              `json:"createdAt"`
}

// WalletHistoryItem represents a wallet transaction history item
type WalletHistoryItem struct {
	TransactionID string    `json:"transactionId"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Direction     string    `json:"direction"`
	Description   string    `json:"description"`
	Reference     string    `json:"reference"`
	CreatedAt     time.Time `json:"createdAt"`
}
