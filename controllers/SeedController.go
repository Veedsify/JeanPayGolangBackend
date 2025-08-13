package controllers

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/gin-gonic/gin"
)

// SeedExchangeRatesEndpoint manually seeds exchange rates
func SeedExchangeRatesEndpoint(c *gin.Context) {
	// Seed exchange rates
	database.SeedExchangeRates()

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Exchange rates seeded successfully",
		"data": gin.H{
			"seeded": []string{"NGN-GHS", "GHS-NGN"},
		},
	})
}

// SeedAllDefaultDataEndpoint seeds all default data
func SeedAllDefaultDataEndpoint(c *gin.Context) {
	// Seed all default data
	database.SeedDefaultData()

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "All default data seeded successfully",
		"data": gin.H{
			"seeded": []string{"exchange_rates"},
		},
	})
}

// SeedTestDataEndpoint seeds test data for development
func SeedTestDataEndpoint(c *gin.Context) {
	// Seed test data
	database.SeedTestData()

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Test data seeded successfully",
		"data": gin.H{
			"environment": "development",
		},
	})
}
