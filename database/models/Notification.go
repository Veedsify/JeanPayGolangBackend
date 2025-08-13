package models

import (
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	UserID  uint   `json:"user_id" gorm:"not null;uniqueIndex"`
	Type    string `json:"type" gorm:"not null"`
	Message string `json:"message" gorm:"not null"`
	Read    bool   `json:"read" gorm:"default:false"`
}

func (Notification) TableName() string {
	return "notifications"
}
