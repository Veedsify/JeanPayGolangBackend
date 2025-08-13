package models

import (
	"time"

	"gorm.io/gorm"
)

type ExchangeRateSource string

const (
	Manual ExchangeRateSource = "manual"
	API    ExchangeRateSource = "api"
)

type ExchangeRate struct {
	gorm.Model
	FromCurrency string             `json:"from_currency" gorm:"not null"`
	ToCurrency   string             `json:"to_currency" gorm:"not null"`
	Rate         float64            `json:"rate" gorm:"not null"`
	Source       ExchangeRateSource `json:"source" gorm:"default:api"`
	SetBy        string             `json:"set_by"` // adminId
	IsActive     bool               `json:"is_active" gorm:"default:true"`
	ValidFrom    time.Time          `json:"valid_from" gorm:"default:CURRENT_TIMESTAMP"`
	ValidTo      *time.Time         `json:"valid_to"`
}

func (ExchangeRate) TableName() string {
	return "exchange_rates"
}
