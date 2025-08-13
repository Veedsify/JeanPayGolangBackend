package models

import (
	"gorm.io/gorm"
)

type Rate struct {
	gorm.Model
	FromCurrency string  `json:"from_currency" gorm:"not null"`
	ToCurrency   string  `json:"to_currency" gorm:"not null"`
	Rate         float64 `json:"rate" gorm:"not null"`
	Source       string  `json:"source" gorm:"not null"`
	Active       bool    `json:"active" gorm:"default:true"`
}

func (Rate) TableName() string {
	return "rates"
}
