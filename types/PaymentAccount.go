package types

import (
	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

// CreatePaymentAccountRequest represents the request to create a new payment account
type CreatePaymentAccountRequest struct {
	Currency      string `json:"currency" binding:"required,oneof=NGN GHS"`
	AccountType   string `json:"account_type" binding:"required,oneof=bank momo"`
	AccountName   string `json:"account_name" binding:"required,min=2,max=100"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BankCode      string `json:"bank_code"`
	PhoneNumber   string `json:"phone_number"`
	Network       string `json:"network"`
	Notes         string `json:"notes"`
}

// UpdatePaymentAccountRequest represents the request to update an existing payment account
type UpdatePaymentAccountRequest struct {
	Currency      *string `json:"currency,omitempty"`
	AccountType   *string `json:"account_type,omitempty"`
	AccountName   *string `json:"account_name,omitempty"`
	AccountNumber *string `json:"account_number,omitempty"`
	BankName      *string `json:"bank_name,omitempty"`
	BankCode      *string `json:"bank_code,omitempty"`
	PhoneNumber   *string `json:"phone_number,omitempty"`
	Network       *string `json:"network,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// PaymentAccountResponse represents the response format for payment account data
type PaymentAccountResponse struct {
	ID            uint               `json:"id"`
	Currency      string             `json:"currency"`
	AccountType   models.AccountType `json:"account_type"`
	AccountName   string             `json:"account_name"`
	AccountNumber string             `json:"account_number,omitempty"`
	BankName      string             `json:"bank_name,omitempty"`
	BankCode      string             `json:"bank_code,omitempty"`
	PhoneNumber   string             `json:"phone_number,omitempty"`
	Network       string             `json:"network,omitempty"`
	IsActive      bool               `json:"is_active"`
	CreatedBy     string             `json:"created_by"`
	Notes         string             `json:"notes"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}

// PaymentAccountsResponse represents paginated payment accounts response
type PaymentAccountsResponse struct {
	Accounts   []PaymentAccountResponse `json:"accounts"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPages int64                    `json:"total_pages"`
}

// ActivePaymentAccountResponse represents simplified response for active accounts used in topup
type ActivePaymentAccountResponse struct {
	ID            uint   `json:"id"`
	Currency      string `json:"currency"`
	AccountType   string `json:"account_type"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	BankCode      string `json:"bank_code,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Network       string `json:"network,omitempty"`
}
