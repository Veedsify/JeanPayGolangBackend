package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateExchangeRateRequest struct {
	FromCurrency string                    `json:"from_currency" form:"from_currency" binding:"required"`
	ToCurrency   string                    `json:"to_currency" form:"to_currency" binding:"required"`
	Rate         float64                   `json:"rate" form:"rate" binding:"required"`
	Source       models.ExchangeRateSource `json:"source" form:"source"`
	SetBy        string                    `json:"set_by" form:"set_by"`
	ValidFrom    time.Time                 `json:"valid_from" form:"valid_from"`
	ValidTo      *time.Time                `json:"valid_to" form:"valid_to"`
}

type UpdateExchangeRateRequest struct {
	Rate      float64                   `json:"rate" form:"rate"`
	Source    models.ExchangeRateSource `json:"source" form:"source"`
	SetBy     string                    `json:"set_by" form:"set_by"`
	IsActive  *bool                     `json:"is_active" form:"is_active"`
	ValidFrom time.Time                 `json:"valid_from" form:"valid_from"`
	ValidTo   *time.Time                `json:"valid_to" form:"valid_to"`
}

type ExchangeRateResponse struct {
	ID           uint32                    `json:"id"`
	FromCurrency string                    `json:"from_currency"`
	ToCurrency   string                    `json:"to_currency"`
	Rate         float64                   `json:"rate"`
	Source       models.ExchangeRateSource `json:"source"`
	SetBy        string                    `json:"set_by"`
	IsActive     bool                      `json:"is_active"`
	ValidFrom    time.Time                 `json:"valid_from"`
	ValidTo      *time.Time                `json:"valid_to"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

type GetExchangeRatesRequest struct {
	FromCurrency string `form:"from_currency"`
	ToCurrency   string `form:"to_currency"`
	Source       string `form:"source"`
	IsActive     *bool  `form:"is_active"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type GetExchangeRatesResponse struct {
	Rates      []ExchangeRateResponse `json:"rates"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

func ToExchangeRateResponse(rate *models.ExchangeRate) ExchangeRateResponse {
	return ExchangeRateResponse{
		ID:           uint32(rate.ID),
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Source:       rate.Source,
		SetBy:        rate.SetBy,
		IsActive:     rate.IsActive,
		ValidFrom:    rate.ValidFrom,
		ValidTo:      rate.ValidTo,
		CreatedAt:    rate.CreatedAt,
		UpdatedAt:    rate.UpdatedAt,
	}
}

func ToExchangeRatesResponse(rates []models.ExchangeRate) []ExchangeRateResponse {
	var response []ExchangeRateResponse
	for _, rate := range rates {
		response = append(response, ToExchangeRateResponse(&rate))
	}
	return response
}
