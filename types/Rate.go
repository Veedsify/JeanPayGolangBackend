package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateRateRequest struct {
	FromCurrency string  `json:"from_currency" form:"from_currency" binding:"required"`
	ToCurrency   string  `json:"to_currency" form:"to_currency" binding:"required"`
	Rate         float64 `json:"rate" form:"rate" binding:"required"`
	Source       string  `json:"source" form:"source" binding:"required"`
}

type UpdateRateRequest struct {
	Rate   float64 `json:"rate" form:"rate"`
	Source string  `json:"source" form:"source"`
	Active *bool   `json:"active" form:"active"`
}

type RateResponse struct {
	ID           uint32    `json:"id"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	Source       string    `json:"source"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GetRatesRequest struct {
	FromCurrency string `form:"from_currency"`
	ToCurrency   string `form:"to_currency"`
	Source       string `form:"source"`
	Active       *bool  `form:"active"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type GetRatesResponse struct {
	Rates      []RateResponse `json:"rates"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

type GetCurrentRateRequest struct {
	FromCurrency string `form:"from_currency" binding:"required"`
	ToCurrency   string `form:"to_currency" binding:"required"`
}

type GetCurrentRateResponse struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	Source       string    `json:"source"`
	LastUpdated  time.Time `json:"last_updated"`
	IsActive     bool      `json:"is_active"`
}

type RateComparisonRequest struct {
	FromCurrency string   `form:"from_currency" binding:"required"`
	ToCurrency   string   `form:"to_currency" binding:"required"`
	Sources      []string `form:"sources"`
}

type RateComparisonResponse struct {
	FromCurrency string                       `json:"from_currency"`
	ToCurrency   string                       `json:"to_currency"`
	Rates        []RateComparisonItemResponse `json:"rates"`
	BestRate     RateComparisonItemResponse   `json:"best_rate"`
	Timestamp    time.Time                    `json:"timestamp"`
}

type RateComparisonItemResponse struct {
	Source      string    `json:"source"`
	Rate        float64   `json:"rate"`
	LastUpdated time.Time `json:"last_updated"`
	IsActive    bool      `json:"is_active"`
}

type RateHistoryRequest struct {
	FromCurrency string `form:"from_currency" binding:"required"`
	ToCurrency   string `form:"to_currency" binding:"required"`
	Source       string `form:"source"`
	FromDate     string `form:"from_date"`
	ToDate       string `form:"to_date"`
	Interval     string `form:"interval"` // "daily", "weekly", "monthly"
}

type RateHistoryResponse struct {
	FromCurrency string                    `json:"from_currency"`
	ToCurrency   string                    `json:"to_currency"`
	Source       string                    `json:"source"`
	Interval     string                    `json:"interval"`
	History      []RateHistoryItemResponse `json:"history"`
}

type RateHistoryItemResponse struct {
	Date        time.Time `json:"date"`
	Rate        float64   `json:"rate"`
	HighRate    float64   `json:"high_rate"`
	LowRate     float64   `json:"low_rate"`
	AvgRate     float64   `json:"avg_rate"`
	ChangePct   float64   `json:"change_pct"`
	ChangeValue float64   `json:"change_value"`
}

func ToRateResponse(rate *models.Rate) RateResponse {
	return RateResponse{
		ID:           uint32(rate.ID),
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Source:       rate.Source,
		Active:       rate.Active,
		CreatedAt:    rate.CreatedAt,
		UpdatedAt:    rate.UpdatedAt,
	}
}

func ToRatesResponse(rates []models.Rate) []RateResponse {
	var response []RateResponse
	for _, rate := range rates {
		response = append(response, ToRateResponse(&rate))
	}
	return response
}
