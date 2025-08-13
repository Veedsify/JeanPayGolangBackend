package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateTransactionRequest struct {
	UserID          string                      `json:"user_id" form:"user_id" binding:"required"`
	Amount          float64                     `json:"amount" form:"amount" binding:"required"`
	Currency        string                      `json:"currency" form:"currency" binding:"required"`
	TransactionType models.TransactionType      `json:"transaction_type" form:"transaction_type" binding:"required"`
	Reference       string                      `json:"reference" form:"reference" binding:"required"`
	Direction       models.TransactionDirection `json:"direction" form:"direction" binding:"required"`
	Description     string                      `json:"description" form:"description"`
}

type UpdateTransactionRequest struct {
	Status      models.TransactionStatus `json:"status" form:"status"`
	Description string                   `json:"description" form:"description"`
}

type GetTransactionsRequest struct {
	UserID          string `form:"user_id"`
	Currency        string `form:"currency"`
	Status          string `form:"status"`
	TransactionType string `form:"transaction_type"`
	Direction       string `form:"direction"`
	FromDate        string `form:"from_date"`
	ToDate          string `form:"to_date"`
	Page            int    `form:"page"`
	Limit           int    `form:"limit"`
}

type GetTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
	TotalPages   int                   `json:"total_pages"`
}

type TransactionStatsRequest struct {
	UserID   string `form:"user_id"`
	FromDate string `form:"from_date"`
	ToDate   string `form:"to_date"`
}

type TransactionStatsResponse struct {
	TotalTransactions  int64                 `json:"total_transactions"`
	TotalVolume        float64               `json:"total_volume"`
	PendingCount       int64                 `json:"pending_count"`
	CompletedCount     int64                 `json:"completed_count"`
	FailedCount        int64                 `json:"failed_count"`
	ByStatus           map[string]int64      `json:"by_status"`
	ByType             map[string]int64      `json:"by_type"`
	ByDirection        map[string]int64      `json:"by_direction"`
	ByCurrency         map[string]float64    `json:"by_currency"`
	RecentActivity     []TransactionResponse `json:"recent_activity"`
	TotalDeposits      float64               `json:"total_deposits"`
	TotalWithdrawals   float64               `json:"total_withdrawals"`
	TotalConversions   float64               `json:"total_conversions"`
	MonthlyDeposits    float64               `json:"monthly_deposits"`
	MonthlyWithdrawals float64               `json:"monthly_withdrawals"`
	MonthlyConversions float64               `json:"monthly_conversions"`
	MonthlyFees        float64               `json:"monthly_fees"`
}

type VerifyTransactionRequest struct {
	TransactionID string `json:"transaction_id" form:"transaction_id" binding:"required"`
	Reference     string `json:"reference" form:"reference" binding:"required"`
}

type VerifyTransactionResponse struct {
	TransactionID string                   `json:"transaction_id"`
	Reference     string                   `json:"reference"`
	Status        models.TransactionStatus `json:"status"`
	Amount        float64                  `json:"amount"`
	Currency      string                   `json:"currency"`
	IsValid       bool                     `json:"is_valid"`
	VerifiedAt    time.Time                `json:"verified_at"`
}

func ToTransactionResponse(transaction *models.Transaction) TransactionResponse {
	return TransactionResponse{
		UserID:          transaction.UserID,
		TransactionID:   transaction.TransactionID,
		Status:          transaction.Status,
		TransactionType: transaction.TransactionType,
		Reference:       transaction.Reference,
		Direction:       transaction.Direction,
		Description:     transaction.Description,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}
}

func ToTransactionsResponse(transactions []models.Transaction) []TransactionResponse {
	var response []TransactionResponse
	for _, transaction := range transactions {
		response = append(response, ToTransactionResponse(&transaction))
	}
	return response
}

// Additional transaction types for the endpoints

type TransactionFilterRequest struct {
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	Currency  string    `json:"currency"`
	Direction string    `json:"direction"`
	MinAmount float64   `json:"min_amount"`
	MaxAmount float64   `json:"max_amount"`
	FromDate  time.Time `json:"from_date"`
	ToDate    time.Time `json:"to_date"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
}

type TransactionStatusHistory struct {
	Status     string    `json:"status"`
	UpdatedBy  uint32    `json:"updated_by"`
	UpdatedAt  time.Time `json:"updated_at"`
	Reason     string    `json:"reason"`
	AdminNotes string    `json:"admin_notes"`
}

type NewTransactionRequest struct {
	FromCurrency    string  `json:"fromCurrency"`
	ToCurrency      string  `json:"toCurrency"`
	FromAmount      string  `json:"fromAmount"`
	ToAmount        string  `json:"toAmount"`
	ExchangeRate    float64 `json:"exchangeRate"`
	Method          string  `json:"method"`
	AccountNumber   string  `json:"accountNumber"`
	BankCode        string  `json:"bankCode"`
	BankName        string  `json:"bankName"`
	PhoneNumber     string  `json:"phoneNumber"`
	Network         string  `json:"network"`
	RecipientName   string  `json:"recipientName"`
	TransactionId   string  `json:"transactionId"`
	MethodOfPayment string  `json:"method_of_payment"`
	CreatedAt       string  `json:"created_at"`
}

// TransactionFilter represents filters for transaction queries
type TransactionFilter struct {
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
	Currency string `json:"currency,omitempty"`
	FromDate string `json:"fromDate,omitempty"`
	ToDate   string `json:"toDate,omitempty"`
	UserID   uint32 `json:"userId,omitempty"`
	Search   string `json:"search,omitempty"`
}

// TransactionResponse represents transaction data for API responses
type TransactionResponse struct {
	ID                 uint                        `json:"id"`
	TransactionID      string                      `json:"transactionId"`
	UserID             uint                        `json:"userId"`
	Amount             float64                     `json:"amount"`
	Currency           string                      `json:"currency"`
	Status             models.TransactionStatus    `json:"status"`
	TransactionType    models.TransactionType      `json:"transactionType"`
	Reference          string                      `json:"reference"`
	Direction          models.TransactionDirection `json:"direction"`
	Description        string                      `json:"description"`
	CreatedAt          time.Time                   `json:"createdAt"`
	UpdatedAt          time.Time                   `json:"updatedAt"`
	User               *UserInfo                   `json:"user,omitempty"`
	TransactionDetails *models.TransactionDetails  `json:"transactionDetails,omitempty"`
}
type CreateNewTransactionResponse struct {
	Transaction    TransactionResponse `json:"transaction"`
	RedirectionURL string              `json:"redirectionUrl"`
	ShouldRedirect bool                `json:"shouldRedirect"`
}

// UserInfo represents basic user information for transaction responses
type UserInfo struct {
	UserID    uint32 `json:"user_id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}
