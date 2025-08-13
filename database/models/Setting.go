package models

import (
	"time"

	"gorm.io/gorm"
)

type DefaultCurrency string

const (
	CEDIS DefaultCurrency = "GHS"
	NAIRA DefaultCurrency = "NGN"
)

type Setting struct {
	gorm.Model
	UserID          uint32          `json:"user_id" gorm:"not null;index"`
	DefaultCurrency DefaultCurrency `json:"default_currency" gorm:"default:'NGN'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Setting) TableName() string {
	return "settings"
}
