package models

import (
	"gorm.io/gorm"
)

type WithdrawMethod struct {
	gorm.Model
	UserID        uint   `json:"user_id" gorm:"not null;index"`
	Currency      string `json:"currency" gorm:"not null;"`
	Method        string `json:"method" gorm:"not null;"`
	AccountNumber string `json:"account_number" gorm:"not null;"`
	AccountName   string `json:"account_name" gorm:"not null;"`
	BankName      string `json:"bank_name" gorm:"not null;"`
	BankCode      string `json:"bank_code" gorm:"not null;"`
}

func TableName() string {
	return "withdrawal_methods"
}
