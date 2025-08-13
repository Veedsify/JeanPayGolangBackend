package models

import (
	"gorm.io/gorm"
)

type NotificationType string

const (
	TransferType NotificationType = "transfer"
	TopUpType    NotificationType = "topup"
	WithdrawType NotificationType = "withdraw"
)

type Notification struct {
	gorm.Model
	UserID  uint             `json:"user_id" gorm:"not null;uniqueIndex"`
	Type    NotificationType `json:"type" gorm:"not null"`
	Title   string           `json:"title" gorm:"not null"`
	Message string           `json:"message" gorm:"not null"`
	Read    bool             `json:"read" gorm:"default:false"`
}

func (Notification) TableName() string {
	return "notifications"
}
