package types

import "time"

// WithdrawalQueryParams represents query parameters for withdrawal requests
type WithdrawalQueryParams struct {
	Page      int      `json:"page"`
	Limit     int      `json:"limit"`
	Status    string   `json:"status"`
	Currency  string   `json:"currency"`
	Method    string   `json:"method"`
	Search    string   `json:"search"`
	FromDate  string   `json:"from_date"`
	ToDate    string   `json:"to_date"`
	MinAmount *float64 `json:"min_amount"`
	MaxAmount *float64 `json:"max_amount"`
}

// WithdrawalAccountDetails represents account details for withdrawal
type WithdrawalAccountDetails struct {
	AccountNumber string `json:"accountNumber,omitempty"`
	BankCode      string `json:"bankCode,omitempty"`
	BankName      string `json:"bankName,omitempty"`
	AccountName   string `json:"accountName,omitempty"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
	Network       string `json:"network,omitempty"`
	RecipientName string `json:"recipientName,omitempty"`
}

// WithdrawalUser represents user information for withdrawal
type WithdrawalUser struct {
	UserID             int     `json:"user_id"`
	FirstName          string  `json:"first_name"`
	LastName           string  `json:"last_name"`
	Email              string  `json:"email"`
	PhoneNumber        string  `json:"phone_number"`
	CountryCode        string  `json:"country_code"`
	ProfilePicture     *string `json:"profile_picture,omitempty"`
	VerificationStatus string  `json:"verification_status"`
}

// WithdrawalRequest represents a withdrawal request
type WithdrawalRequest struct {
	ID              int                      `json:"id"`
	TransactionID   string                   `json:"transaction_id"`
	UserID          int                      `json:"user_id"`
	Amount          float64                  `json:"amount"`
	Currency        string                   `json:"currency"`
	Fee             float64                  `json:"fee"`
	NetAmount       float64                  `json:"net_amount"`
	Status          string                   `json:"status"`
	Method          string                   `json:"method"`
	AccountDetails  WithdrawalAccountDetails `json:"account_details"`
	ReceiptURL      *string                  `json:"receipt_url,omitempty"`
	AdminNotes      string                   `json:"admin_notes,omitempty"`
	RejectionReason string                   `json:"rejection_reason,omitempty"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
	ProcessedAt     *string                  `json:"processed_at,omitempty"`
	ProcessedBy     *int                     `json:"processed_by,omitempty"`
}

// WithdrawalWithUser represents a withdrawal request with user information
type WithdrawalWithUser struct {
	ID              int                      `json:"id"`
	TransactionID   string                   `json:"transaction_id"`
	UserID          int                      `json:"user_id"`
	Amount          float64                  `json:"amount"`
	Currency        string                   `json:"currency"`
	Fee             float64                  `json:"fee"`
	NetAmount       float64                  `json:"net_amount"`
	Status          string                   `json:"status"`
	Method          string                   `json:"method"`
	AccountDetails  WithdrawalAccountDetails `json:"account_details"`
	ReceiptURL      *string                  `json:"receipt_url,omitempty"`
	AdminNotes      string                   `json:"admin_notes,omitempty"`
	RejectionReason string                   `json:"rejection_reason,omitempty"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
	ProcessedAt     *string                  `json:"processed_at,omitempty"`
	ProcessedBy     *int                     `json:"processed_by,omitempty"`
	User            WithdrawalUser           `json:"user"`
}

// WithdrawalPeriodStats represents statistics for a specific period
type WithdrawalPeriodStats struct {
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

// WithdrawalStats represents withdrawal statistics
type WithdrawalStats struct {
	TotalRequests     int64                 `json:"total_requests"`
	PendingRequests   int64                 `json:"pending_requests"`
	CompletedRequests int64                 `json:"completed_requests"`
	FailedRequests    int64                 `json:"failed_requests"`
	TotalAmount       float64               `json:"total_amount"`
	TotalFees         float64               `json:"total_fees"`
	Today             WithdrawalPeriodStats `json:"today"`
	ThisWeek          WithdrawalPeriodStats `json:"this_week"`
	ThisMonth         WithdrawalPeriodStats `json:"this_month"`
}

// WithdrawalPagination represents pagination information for withdrawals
type WithdrawalPagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// WithdrawalActionResponse represents the response for withdrawal actions
type WithdrawalActionResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
	ActionedBy *int   `json:"actioned_by,omitempty"`
}

// ApproveWithdrawalRequest represents a request to approve a withdrawal
type ApproveWithdrawalRequest struct {
	AdminNotes string `json:"admin_notes,omitempty"`
}

// RejectWithdrawalRequest represents a request to reject a withdrawal
type RejectWithdrawalRequest struct {
	Reason     string `json:"reason" binding:"required"`
	AdminNotes string `json:"admin_notes,omitempty"`
}

// UpdateWithdrawalStatusRequest represents a request to update withdrawal status
type UpdateWithdrawalStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	Reason     string `json:"reason,omitempty"`
	AdminNotes string `json:"admin_notes,omitempty"`
}

// WithdrawalStatusHistory represents status change history for a withdrawal
type WithdrawalStatusHistory struct {
	ID           int       `json:"id"`
	WithdrawalID string    `json:"withdrawal_id"`
	Status       string    `json:"status"`
	UpdatedBy    int       `json:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at"`
	Reason       *string   `json:"reason,omitempty"`
	AdminNotes   *string   `json:"admin_notes,omitempty"`
	Admin        struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	} `json:"admin"`
}

// WithdrawalNote represents an admin note for a withdrawal
type WithdrawalNote struct {
	ID           int       `json:"id"`
	WithdrawalID string    `json:"withdrawal_id"`
	AdminID      int       `json:"admin_id"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
	Admin        struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	} `json:"admin"`
}

// FeeCalculation represents fee calculation details
type FeeCalculation struct {
	Amount        float64  `json:"amount"`
	Currency      string   `json:"currency"`
	FeeRate       float64  `json:"fee_rate"`
	CalculatedFee float64  `json:"calculated_fee"`
	FeeCap        *float64 `json:"fee_cap,omitempty"`
	FinalFee      float64  `json:"final_fee"`
	NetAmount     float64  `json:"net_amount"`
}

// WithdrawalDetailsResponse represents detailed withdrawal information
type WithdrawalDetailsResponse struct {
	ID              int                       `json:"id"`
	TransactionID   string                    `json:"transaction_id"`
	UserID          int                       `json:"user_id"`
	Amount          float64                   `json:"amount"`
	Currency        string                    `json:"currency"`
	Fee             float64                   `json:"fee"`
	NetAmount       float64                   `json:"net_amount"`
	Status          string                    `json:"status"`
	Method          string                    `json:"method"`
	AccountDetails  WithdrawalAccountDetails  `json:"account_details"`
	ReceiptURL      *string                   `json:"receipt_url,omitempty"`
	AdminNotes      string                    `json:"admin_notes,omitempty"`
	RejectionReason string                    `json:"rejection_reason,omitempty"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
	ProcessedAt     *string                   `json:"processed_at,omitempty"`
	ProcessedBy     *int                      `json:"processed_by,omitempty"`
	User            WithdrawalUser            `json:"user"`
	StatusHistory   []WithdrawalStatusHistory `json:"status_history"`
	AdminNotesList  []WithdrawalNote          `json:"admin_notes_list"`
	FeeCalculation  FeeCalculation            `json:"fee_calculation"`
}

// BulkWithdrawalAction represents a bulk action on withdrawals
type BulkWithdrawalAction struct {
	Action        string   `json:"action" binding:"required"`
	WithdrawalIDs []string `json:"withdrawal_ids" binding:"required"`
	Reason        string   `json:"reason,omitempty"`
	AdminNotes    string   `json:"admin_notes,omitempty"`
}

// BulkWithdrawalResponse represents the response for bulk withdrawal actions
type BulkWithdrawalResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Processed int    `json:"processed"`
	Failed    int    `json:"failed"`
	Errors    []struct {
		WithdrawalID string `json:"withdrawal_id"`
		Error        string `json:"error"`
	} `json:"errors"`
}
