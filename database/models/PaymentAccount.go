package models

import (
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeBank AccountType = "bank"
	AccountTypeMomo AccountType = "momo"
)

type PaymentAccount struct {
	gorm.Model
	Currency      string      `json:"currency" gorm:"not null"`      // NGN or GHS
	AccountType   AccountType `json:"account_type" gorm:"not null"`  // bank or momo
	AccountName   string      `json:"account_name" gorm:"not null"`  // Account holder name
	AccountNumber string      `json:"account_number" gorm:"null"`    // Bank account number (for bank type)
	BankName      string      `json:"bank_name" gorm:"null"`         // Bank name (for bank type)
	BankCode      string      `json:"bank_code" gorm:"null"`         // Bank code (for bank type)
	PhoneNumber   string      `json:"phone_number" gorm:"null"`      // Phone number (for momo type)
	Network       string      `json:"network" gorm:"null"`           // Network provider (for momo type)
	IsActive      bool        `json:"is_active" gorm:"default:true"` // Whether account is active
	CreatedBy     uint        `json:"created_by" gorm:"not null"`    // Admin ID who created this
	Notes         string      `json:"notes" gorm:"type:text"`        // Additional notes
}

func (PaymentAccount) TableName() string {
	return "payment_accounts"
}
