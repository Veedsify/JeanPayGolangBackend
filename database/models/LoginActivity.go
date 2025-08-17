package models

import (
	"gorm.io/gorm"
)

type LoginActivity struct {
	gorm.Model
	UserID   uint   `json:"user_id" gorm:"not null;index"`
	Activity string `json:"activity" gorm:"not null"`
}

func (LoginActivity) TableName() string {
	return "login_activities"
}
