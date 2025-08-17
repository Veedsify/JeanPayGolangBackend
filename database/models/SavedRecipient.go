package models

import "gorm.io/gorm"

type RecipientType string

const (
	RecipientTypeBank RecipientType = "bank"
	RecipientTypeMomo RecipientType = "momo"
)

type SavedRecipient struct {
	*gorm.Model
	UserID            uint   `json:"user_id" gorm:"not null;index"`
	RecipientName     string `json:"recipient_name" gorm:"not null"`
	RecipientPhone    string `json:"recipient_phone" gorm:"not null"`
	RecipientAccount  string `json:"recipient_account" gorm:"not null"`
	RecipientBank     string `json:"recipient_bank" gorm:"not null"`
	RecipientBankCode string `json:"recipient_bank_code" gorm:"not null"`
	RecipientType     string `json:"recipient_type" gorm:"not null"` // e.g momo or bank transfer
}

func (SavedRecipient) TableName() string {
	return "saved_recipients"
}
