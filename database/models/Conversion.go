package models

import (
	"gorm.io/gorm"
)

type ConversionStatus string

const (
	ConversionPending   ConversionStatus = "pending"
	ConversionCompleted ConversionStatus = "completed"
	ConversionFailed    ConversionStatus = "failed"
)

type Conversions struct {
	gorm.Model
	UserID           uint             `json:"user_id" gorm:"not null;uniqueIndex"`
	ConversionID     string           `json:"conversion_id" gorm:"not null;uniqueIndex"`
	TransactionID    string           `json:"transaction_id" gorm:"not null"`
	FromCurrency     string           `json:"from_currency" gorm:"not null"`
	ToCurrency       string           `json:"to_currency" gorm:"not null"`
	Amount           float64          `json:"amount" gorm:"not null"`
	ConvertedAmount  float64          `json:"converted_amount" gorm:"not null"`
	Fee              float64          `json:"fee" gorm:"not null"`
	Rate             float64          `json:"rate" gorm:"not null"`
	Source           string           `json:"source" gorm:"not null"`
	Status           ConversionStatus `json:"status" gorm:"default:pending"`
	EstimatedArrival string           `json:"estimated_arrival"`
}

func (Conversions) TableName() string {
	return "conversions"
}
