package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateConversionRequest struct {
	UserID           string  `json:"user_id" form:"user_id" binding:"required"`
	TransactionID    string  `json:"transaction_id" form:"transaction_id" binding:"required"`
	FromCurrency     string  `json:"from_currency" form:"from_currency" binding:"required"`
	ToCurrency       string  `json:"to_currency" form:"to_currency" binding:"required"`
	Amount           float64 `json:"amount" form:"amount" binding:"required"`
	ConvertedAmount  float64 `json:"converted_amount" form:"converted_amount" binding:"required"`
	Fee              float64 `json:"fee" form:"fee" binding:"required"`
	Rate             float64 `json:"rate" form:"rate" binding:"required"`
	Source           string  `json:"source" form:"source" binding:"required"`
	EstimatedArrival string  `json:"estimated_arrival" form:"estimated_arrival"`
}

type UpdateConversionRequest struct {
	Status           models.ConversionStatus `json:"status" form:"status"`
	ConvertedAmount  float64                 `json:"converted_amount" form:"converted_amount"`
	Fee              float64                 `json:"fee" form:"fee"`
	Rate             float64                 `json:"rate" form:"rate"`
	EstimatedArrival string                  `json:"estimated_arrival" form:"estimated_arrival"`
}

type GetConversionsRequest struct {
	UserID       string `form:"user_id"`
	FromCurrency string `form:"from_currency"`
	ToCurrency   string `form:"to_currency"`
	Status       string `form:"status"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type GetConversionsResponse struct {
	Conversions []ConversionResponse `json:"conversions"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	TotalPages  int                  `json:"total_pages"`
}

type ConversionCalculateRequest struct {
	FromCurrency string  `json:"from_currency" form:"from_currency" binding:"required"`
	ToCurrency   string  `json:"to_currency" form:"to_currency" binding:"required"`
	Amount       float64 `json:"amount" form:"amount" binding:"required"`
}

type ConversionCalculateResponse struct {
	FromCurrency     string  `json:"from_currency"`
	ToCurrency       string  `json:"to_currency"`
	Amount           float64 `json:"amount"`
	ConvertedAmount  float64 `json:"converted_amount"`
	Fee              float64 `json:"fee"`
	Rate             float64 `json:"rate"`
	EstimatedArrival string  `json:"estimated_arrival"`
}

// ConversionRequest represents a currency conversion request
type ConversionRequest struct {
	FromCurrency string  `json:"fromCurrency" validate:"required,oneof=NGN GHS"`
	ToCurrency   string  `json:"toCurrency" validate:"required,oneof=NGN GHS"`
	Amount       float64 `json:"amount" validate:"required,gt=0"`
}

// ConversionResponse represents the response for a conversion request
type ConversionResponse struct {
	ConversionID     string    `json:"conversionId"`
	TransactionID    string    `json:"transactionId"`
	FromCurrency     string    `json:"fromCurrency"`
	ToCurrency       string    `json:"toCurrency"`
	OriginalAmount   float64   `json:"originalAmount"`
	Fee              float64   `json:"fee"`
	ConvertedAmount  float64   `json:"convertedAmount"`
	Rate             float64   `json:"rate"`
	Status           string    `json:"status"`
	EstimatedArrival string    `json:"estimatedArrival"`
	CreatedAt        time.Time `json:"createdAt"`
}

// ExchangeRatesResponse represents exchange rates response
type ExchangeRatesResponse struct {
	Rates       map[string]float64 `json:"rates"`
	LastUpdated time.Time          `json:"lastUpdated"`
	Source      string             `json:"source"`
}

// CalculationResponse represents conversion calculation response
type CalculationResponse struct {
	FromCurrency     string  `json:"fromCurrency"`
	ToCurrency       string  `json:"toCurrency"`
	OriginalAmount   float64 `json:"originalAmount"`
	Fee              float64 `json:"fee"`
	AmountAfterFee   float64 `json:"amountAfterFee"`
	ConvertedAmount  float64 `json:"convertedAmount"`
	Rate             float64 `json:"rate"`
	EstimatedArrival string  `json:"estimatedArrival"`
}
