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
	UserID                   uint            `json:"user_id" gorm:"not null;index"`
	DefaultCurrency          DefaultCurrency `json:"default_currency" gorm:"default:'NGN'"`
	Username                 string          `json:"username" gorm:"null;index"`
	FeesBreakdown            bool            `json:"fees_breakdown" gorm:"default:true"`
	SaveRecipient            bool            `json:"save_recipient" gorm:"default:false"`
	EmailNotifications       bool            `json:"email_notifications" gorm:"default:true"`
	PushNotifications        bool            `json:"push_notifications" gorm:"default:true"`
	PromotionalNotifications bool            `json:"promotional_notifications" gorm:"default:true"`
	TwoFactorAuth            bool            `json:"two_factor_auth" gorm:"default:false"`
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (Setting) TableName() string {
	return "settings"
}
