package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConvertCurrency performs actual currency conversion
func ConvertCurrency(userID uint, req types.ConversionRequest) (*types.ConversionResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Validate request with country code if provided
	if req.CountryCode != "" {
		if err := validateConversionRequestWithCountry(req, req.CountryCode); err != nil {
			return nil, err
		}
	} else {
		if err := validateConversionRequest(req); err != nil {
			return nil, err
		}
	}

	// Get user wallets
	wallets, err := findOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to access wallets: %w", err)
	}

	// Check sufficient balance for the source currency
	currentBalance := getBalanceByCurrency(wallets, req.FromCurrency)
	if currentBalance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	// Get current exchange rate
	rate, err := getCurrentExchangeRate(req.FromCurrency, req.ToCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Calculate conversion amounts
	fee := utils.CalculateFee(req.Amount, 2.0) // 2% fee
	amountAfterFee := req.Amount - fee
	convertedAmount := utils.RoundCurrency(amountAfterFee * rate)

	// Start database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	conversionID := uuid.New().String()
	transactionID := uuid.New().String()

	// Create conversion record
	conversion := models.Conversions{
		UserID:           userID,
		ConversionID:     conversionID,
		TransactionID:    transactionID,
		FromCurrency:     req.FromCurrency,
		ToCurrency:       req.ToCurrency,
		Amount:           utils.RoundCurrency(req.Amount),
		ConvertedAmount:  convertedAmount,
		Fee:              utils.RoundCurrency(fee),
		Rate:             rate,
		Source:           "user_request",
		Status:           "pending",
		EstimatedArrival: "Instant",
	}

	if err := tx.Create(&conversion).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create conversion: %w", err)
	}

	// Create transaction record
	transaction := models.Transaction{
		UserID:          userID,
		TransactionID:   transactionID,
		Status:          "pending",
		TransactionType: "conversion",
		Reference:       conversionID,
		Direction:       getConversionDirection(req.FromCurrency, req.ToCurrency),
		Description:     fmt.Sprintf("Convert %s %.2f to %s", req.FromCurrency, req.Amount, req.ToCurrency),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update wallet balances
	// Deduct from source currency
	err = updateWalletBalanceInTx(tx, userID, req.FromCurrency, -req.Amount, "conversion")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to deduct from wallet: %w", err)
	}

	// Add to target currency
	err = updateWalletBalanceInTx(tx, userID, req.ToCurrency, convertedAmount, "conversion")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add to wallet: %w", err)
	}

	// Update conversion and transaction status to completed
	if err := tx.Model(&conversion).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update conversion status: %w", err)
	}

	if err := tx.Model(&transaction).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &types.ConversionResponse{
		ConversionID:     conversionID,
		TransactionID:    transactionID,
		FromCurrency:     req.FromCurrency,
		ToCurrency:       req.ToCurrency,
		OriginalAmount:   utils.RoundCurrency(req.Amount),
		Fee:              utils.RoundCurrency(fee),
		ConvertedAmount:  convertedAmount,
		Rate:             rate,
		Status:           "completed",
		EstimatedArrival: "Instant",
		CreatedAt:        now,
	}, nil
}

// GetExchangeRates retrieves current exchange rates
func GetExchangeRates() (*types.ExchangeRatesResponse, error) {
	rates := make(map[string]float64)
	var lastUpdated time.Time
	source := "default"

	// Get all supported currency pairs
	supportedPairs := utils.GetSupportedCurrencyPairs()

	for _, pair := range supportedPairs {
		pairKey := pair.String()
		rate, err := getCurrentExchangeRate(pair.From, pair.To)
		if err == nil {
			rates[pairKey] = rate
			source = "database"
		} else {
			// Use default rates if not in database
			defaultRates := map[string]float64{
				"NGN-GHS": 0.0053,
				"GHS-NGN": 188.68,
				"USD-GHS": 16.50,
				"NGN-XOF": 0.35,
				"GHS-XOF": 66.00,
				"XOF-GHS": 0.015,
				"XOF-NGN": 2.86,
			}

			if defaultRate, exists := defaultRates[pairKey]; exists {
				rates[pairKey] = defaultRate
			}
		}
	}

	lastUpdated = time.Now()

	return &types.ExchangeRatesResponse{
		Rates:       rates,
		LastUpdated: lastUpdated,
		Source:      source,
	}, nil
}

// CalculateConversion calculates conversion amounts without performing the conversion
func CalculateConversion(req types.ConversionRequest) (*types.CalculationResponse, error) {
	// Validate request with country code if provided
	if req.CountryCode != "" {
		if err := validateConversionRequestWithCountry(req, req.CountryCode); err != nil {
			return nil, err
		}
	} else {
		if err := validateConversionRequest(req); err != nil {
			return nil, err
		}
	}

	// Get current exchange rate
	rate, err := getCurrentExchangeRate(req.FromCurrency, req.ToCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	newAmount := rate * req.Amount

	// Calculate conversion amounts
	convertedAmount := utils.RoundCurrency(newAmount)
	return &types.CalculationResponse{
		FromCurrency:     req.FromCurrency,
		ToCurrency:       req.ToCurrency,
		OriginalAmount:   utils.RoundCurrency(req.Amount),
		ConvertedAmount:  convertedAmount,
		Rate:             rate,
		EstimatedArrival: "Instant",
	}, nil
}

// ExecuteConversion performs actual currency conversion
func ExecuteConversion(userID uint, req types.ConversionRequest) (*types.ConversionResponse, error) {
	return ConvertCurrency(userID, req)
}

// ConversionHistoryItem represents a conversion history item
type ConversionHistoryItem struct {
	ConversionID     string    `json:"conversionId"`
	TransactionID    string    `json:"transactionId"`
	FromCurrency     string    `json:"fromCurrency"`
	ToCurrency       string    `json:"toCurrency"`
	Amount           float64   `json:"amount"`
	ConvertedAmount  float64   `json:"convertedAmount"`
	Fee              float64   `json:"fee"`
	Rate             float64   `json:"rate"`
	Status           string    `json:"status"`
	EstimatedArrival string    `json:"estimatedArrival"`
	CreatedAt        time.Time `json:"createdAt"`
}

// GetConversionHistory retrieves conversion history for a user
func GetConversionHistory(userID uint32, pagination types.PaginationRequest, status, fromCurrency, toCurrency string) ([]ConversionHistoryItem, *types.PaginationResponse, error) {
	if userID == 0 {
		return nil, nil, errors.New("user ID is required")
	}

	// Build query
	query := database.DB.Model(&models.Conversions{}).Where("user_id = ?", userID)

	if status != "" && isValidConversionStatus(status) {
		query = query.Where("status = ?", status)
	}

	if fromCurrency != "" && constants.IsValidCurrency(fromCurrency) {
		query = query.Where("from_currency = ?", fromCurrency)
	}

	if toCurrency != "" && constants.IsValidCurrency(toCurrency) {
		query = query.Where("to_currency = ?", toCurrency)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count conversions: %w", err)
	}

	// Calculate pagination
	page := pagination.Page
	if page <= 0 {
		page = 1
	}
	limit := pagination.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Find conversions
	var conversions []models.Conversions
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&conversions).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find conversions: %w", err)
	}

	// Convert to history items
	var historyItems []ConversionHistoryItem
	for _, conv := range conversions {
		historyItems = append(historyItems, ConversionHistoryItem{
			ConversionID:     conv.ConversionID,
			TransactionID:    conv.TransactionID,
			FromCurrency:     conv.FromCurrency,
			ToCurrency:       conv.ToCurrency,
			Amount:           utils.RoundCurrency(conv.Amount),
			ConvertedAmount:  utils.RoundCurrency(conv.ConvertedAmount),
			Fee:              utils.RoundCurrency(conv.Fee),
			Rate:             conv.Rate,
			Status:           string(conv.Status),
			EstimatedArrival: conv.EstimatedArrival,
			CreatedAt:        conv.CreatedAt,
		})
	}

	// Create pagination response
	paginationResp := &types.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + int64(limit) - 1) / int64(limit),
	}

	return historyItems, paginationResp, nil
}

// Helper functions

// getCurrentExchangeRate gets the current exchange rate between two currencies
func getCurrentExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	var rate models.Rate

	err := database.DB.Where("from_currency = ? AND to_currency = ? AND active = ?",
		fromCurrency, toCurrency, true).Order("id DESC").First(&rate).Error

	if err == nil {
		return rate.Rate, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to query exchange rate: %w", err)
	}

	// Return default rates if not in database for supported pairs
	defaultRates := map[string]float64{
		"NGN-GHS": 0.0053,
		"GHS-NGN": 188.68,
		"USD-GHS": 16.50,
		"NGN-XOF": 0.35,
		"GHS-XOF": 66.00,
		"XOF-GHS": 0.015,
		"XOF-NGN": 2.86,
	}

	pairKey := fmt.Sprintf("%s-%s", fromCurrency, toCurrency)
	if defaultRate, exists := defaultRates[pairKey]; exists {
		return defaultRate, nil
	}

	return 0, errors.New("exchange rate not available for the specified currency pair")
}

// getConversionDirection gets transaction direction for conversion
func getConversionDirection(fromCurrency, toCurrency string) models.TransactionDirection {
	direction := fmt.Sprintf("%s-%s", fromCurrency, toCurrency)
	return models.TransactionDirection(direction)
}

// updateWalletBalanceInTx updates wallet balance within a transaction
func updateWalletBalanceInTx(tx *gorm.DB, userID uint, currency string, amount float64, txType string) error {
	var wallet models.Wallet

	// Find wallet with lock
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return fmt.Errorf("failed to find wallet: %w", err)
	}

	// Update balance based on currency
	roundedAmount := utils.RoundCurrency(amount)
	if currency == wallet.Currency {
		wallet.Balance += roundedAmount
	} else {
		// Handle multi-currency wallet updates
		// For now, we'll update the balance if currencies match
		// In a full implementation, you might want separate wallet records per currency
		if currency == "NGN" && wallet.Currency == "NGN" {
			wallet.Balance += roundedAmount
		} else if currency == "GHS" && wallet.Currency == "GHS" {
			wallet.Balance += roundedAmount
		} else if currency == "XOF" && wallet.Currency == "XOF" {
			wallet.Balance += roundedAmount
		} else if currency == "USD" && wallet.Currency == "USD" {
			wallet.Balance += roundedAmount
		}
	}

	// Update totals for conversion
	if txType == "conversion" {
		if amount < 0 {
			wallet.TotalConversions += -roundedAmount
		}
	}

	// Update timestamps
	now := time.Now()
	wallet.UpdatedAt = now
	wallet.LastTransactionAt = &now

	// Save wallet
	if err := tx.Save(&wallet).Error; err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

// Helper functions for validation
func isValidConversionStatus(status string) bool {
	validStatuses := []string{"pending", "completed", "failed"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// validateConversionRequest validates conversion request
func validateConversionRequest(req types.ConversionRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	// Use the new currency pair validation
	if err := utils.ValidateCurrencyPair(req.FromCurrency, req.ToCurrency); err != nil {
		return err
	}

	return nil
}

// validateConversionRequestWithCountry validates conversion request with country validation for XOF
func validateConversionRequestWithCountry(req types.ConversionRequest, countryCode string) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	// Use the new currency pair validation
	if err := utils.ValidateCurrencyPair(req.FromCurrency, req.ToCurrency); err != nil {
		return err
	}

	// Validate XOF country restrictions
	if err := utils.ValidateXOFCountry(req.FromCurrency, countryCode); err != nil {
		return err
	}

	if err := utils.ValidateXOFCountry(req.ToCurrency, countryCode); err != nil {
		return err
	}

	return nil
}
