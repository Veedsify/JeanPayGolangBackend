package database

import (
	"log"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

// SeedExchangeRates seeds default exchange rates for NGN and GHS
func SeedExchangeRates() {
	log.Println("Seeding default exchange rates...")

	// Default exchange rates
	exchangeRates := []models.ExchangeRate{
		{
			FromCurrency: "NGN",
			ToCurrency:   "GHS",
			Rate:         0.0053,
			Source:       models.Manual,
			SetBy:        "system",
			IsActive:     true,
			ValidFrom:    time.Now(),
		},
		{
			FromCurrency: "GHS",
			ToCurrency:   "NGN",
			Rate:         188.68,
			Source:       models.Manual,
			SetBy:        "system",
			IsActive:     true,
			ValidFrom:    time.Now(),
		},
	}

	for _, rate := range exchangeRates {
		// Check if exchange rate already exists
		var existingRate models.ExchangeRate
		result := DB.Where(
			"from_currency = ? AND to_currency = ? AND is_active = ?",
			rate.FromCurrency,
			rate.ToCurrency,
			true,
		).First(&existingRate)

		if result.Error != nil {
			// Exchange rate doesn't exist, create it
			if err := DB.Create(&rate).Error; err != nil {
				log.Printf("Error creating exchange rate %s to %s: %v", rate.FromCurrency, rate.ToCurrency, err)
			} else {
				log.Printf("Created exchange rate: %s to %s = %.6f", rate.FromCurrency, rate.ToCurrency, rate.Rate)
			}
		} else {
			// Exchange rate exists, update it if the rate is different
			if existingRate.Rate != rate.Rate {
				existingRate.Rate = rate.Rate
				existingRate.UpdatedAt = time.Now()
				if err := DB.Save(&existingRate).Error; err != nil {
					log.Printf("Error updating exchange rate %s to %s: %v", rate.FromCurrency, rate.ToCurrency, err)
				} else {
					log.Printf("Updated exchange rate: %s to %s = %.6f", rate.FromCurrency, rate.ToCurrency, rate.Rate)
				}
			} else {
				log.Printf("Exchange rate %s to %s already exists with correct rate", rate.FromCurrency, rate.ToCurrency)
			}
		}
	}

	log.Println("Exchange rates seeding completed")
}

// SeedDefaultData seeds all default data needed for the application
func SeedDefaultData() {
	log.Println("Starting database seeding...")
	SeedExchangeRates()
	// Add other seeding functions here as needed
	log.Println("Database seeding completed")
}

// SeedTestData seeds test data for development/testing
func SeedTestData() {
	log.Println("Seeding test data...")

	// Add test users, transactions, etc. for development
	// This function can be called in development mode only

	log.Println("Test data seeding completed")
}
