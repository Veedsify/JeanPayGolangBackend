package models

import (
	"gorm.io/gorm"
)

type Activity struct {
	gorm.Model
	UserID   uint   `json:"user_id" gorm:"not null;index"`
	Activity string `json:"activity" gorm:"not null"`
}

func (Activity) TableName() string {
	return "activities"
}
