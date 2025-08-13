package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

// AdminLoginRequest represents admin login request
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AdminLoginResponse represents admin login response
type AdminLoginResponse struct {
	Token string               `json:"token"`
	Admin AdminProfileResponse `json:"admin"`
}

// AdminProfileResponse represents admin profile data
type AdminProfileResponse struct {
	UserID    uint32 `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"isAdmin"`
}

// DashboardSummary represents admin dashboard summary
type DashboardSummary struct {
	TotalUsers        int64 `json:"totalUsers"`
	TotalTransactions int64 `json:"totalTransactions"`
	PendingTxns       int64 `json:"pendingTransactions"`
	CompletedTxns     int64 `json:"completedTransactions"`
	FailedTxns        int64 `json:"failedTransactions"`
}

// MonthlyVolume represents monthly transaction volume
type MonthlyVolume struct {
	Direction string  `json:"direction"`
	Total     float64 `json:"total"`
	Count     int64   `json:"count"`
}

// DashboardResponse represents admin dashboard response
type DashboardResponse struct {
	Summary            DashboardSummary      `json:"summary"`
	MonthlyVolume      []MonthlyVolume       `json:"monthlyVolume"`
	RecentTransactions []TransactionWithUser `json:"recentTransactions"`
}

// TransactionWithUser represents transaction with user info
type TransactionWithUser struct {
	TransactionID      string                    `json:"transaction_id"`
	UserID             uint                      `json:"user_id"`
	Amount             float64                   `json:"amount"`
	Currency           string                    `json:"currency"`
	Status             string                    `json:"status"`
	TransactionType    string                    `json:"transaction_type"`
	Reference          string                    `json:"reference"`
	Direction          string                    `json:"direction"`
	Description        string                    `json:"description"`
	CreatedAt          time.Time                 `json:"created_at"`
	PaymentType        string                    `json:"payment_type"`
	TransactionDetails models.TransactionDetails `json:"transaction_details"`
	User               *UserBasicInfo            `json:"user,omitempty"`
}

// TransactionQueryParams represents query parameters for transactions
type TransactionQueryParams struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
	Type   string `form:"type"`
	Search string `form:"search"`
}

// TransactionsResponse represents paginated transactions response
type TransactionsResponse struct {
	Transactions []TransactionWithUser `json:"transactions"`
	Pagination   PaginationInfo        `json:"pagination"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
	Pages int   `json:"pages"`
}

// TransactionDetailsResponse represents detailed transaction response
type TransactionDetailsResponse struct {
	Transaction models.Transaction `json:"transaction"`
}

// UserQueryParams represents query parameters for users
type UserQueryParams struct {
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
	Status  string `form:"status"`
	Blocked string `form:"blocked"`
	Search  string `form:"search"`
}

// UserWithWallet represents user with wallet information
type UserWithWallet struct {
	User             models.User    `json:"user"`
	Wallet           *WalletBalance `json:"wallet"`
	TransactionCount int64          `json:"transactionCount"`
}

// UsersResponse represents paginated users response
type UsersResponse struct {
	Users      []UserWithWallet `json:"users"`
	Pagination PaginationInfo   `json:"pagination"`
}

// ApproveTransactionRequest represents approve transaction request
type ApproveTransactionRequest struct {
	TransactionID string `json:"transactionId" validate:"required"`
}

// RejectTransactionRequest represents reject transaction request
type RejectTransactionRequest struct {
	TransactionID string `json:"transactionId" validate:"required"`
	Reason        string `json:"reason"`
}

// BlockUserRequest represents block user request
type BlockUserRequest struct {
	UserID string `json:"userId" validate:"required"`
	Reason string `json:"reason"`
}

// UnblockUserRequest represents unblock user request
type UnblockUserRequest struct {
	UserID string `json:"userId" validate:"required"`
}

// AdminActionResponse represents response for admin actions
type AdminActionResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// UserFilter represents user filter criteria
type UserFilter struct {
	Status  string `form:"status"`
	Blocked string `form:"blocked"`
	Search  string `form:"search"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AdminTransactionQuery represents query parameters for admin transactions
type AdminTransactionQuery struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
	Type   string `form:"type"`
	Search string `form:"search"`
}

// UserTransactionQuery represents query parameters for user transactions
type UserTransactionQuery struct {
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	Status      string `form:"status"`
	Type        string `form:"type"`
	Currency    string `form:"currency"`
	FromDate    string `form:"from_date"`
	ToDate      string `form:"to_date"`
	AccountType string `form:"account_type"`
	Search      string `form:"search"`
}

// UpdateTransactionStatusRequest represents request to update transaction status
type UpdateTransactionStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// CreateDepositRequest represents request to create deposit transaction
type CreateDepositRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	PaymentRef  string  `json:"payment_ref" binding:"required"`
	Description string  `json:"description,omitempty"`
}

// CreateWithdrawalRequest represents request to create withdrawal transaction
type CreateWithdrawalRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	BankAccount string  `json:"bank_account" binding:"required"`
	Description string  `json:"description,omitempty"`
}

// TransactionActionResponse represents response for transaction actions
type TransactionActionResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	ActionedAt    string `json:"actioned_at"`
	ActionedBy    uint32 `json:"actioned_by"`
	Reason        string `json:"reason,omitempty"`
}

// AdminTransactionResponse represents paginated admin transactions response
type AdminTransactionResponse struct {
	Transactions []TransactionWithUser `json:"transactions"`
	Pagination   PaginationInfo        `json:"pagination"`
}

// UserBasicInfo represents basic user information
type UserBasicInfo struct {
	UserID      uint   `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

// WalletInfo represents wallet information
type WalletInfo struct {
	UserID           string  `json:"user_id"`
	BalanceNGN       float64 `json:"balance_ngn"`
	BalanceGHS       float64 `json:"balance_ghs"`
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
	TotalConversions float64 `json:"total_conversions"`
	IsActive         bool    `json:"is_active"`
}
