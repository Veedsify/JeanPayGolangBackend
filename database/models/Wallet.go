package models

import (
	"time"

	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	UserID            uint       `json:"user_id" gorm:"not null;index"`
	Currency          string     `json:"currency" gorm:"not null;default:'NGN';enum('NGN', 'GHS')"`
	Balance           float64    `json:"balance" gorm:"default:0"`
	WalletID          uint64     `json:"wallet_id" gorm:"not null;uniqueIndex"`
	TotalDeposits     float64    `json:"total_deposits" gorm:"default:0"`
	TotalWithdrawals  float64    `json:"total_withdrawals" gorm:"default:0"`
	TotalConversions  float64    `json:"total_conversions" gorm:"default:0"`
	IsActive          bool       `json:"is_active" gorm:"default:true"`
	LastTransactionAt *time.Time `json:"last_transaction_at"`
}

func (Wallet) TableName() string {
	return "wallets"
}
