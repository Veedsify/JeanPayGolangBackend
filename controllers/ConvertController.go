package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// GetExchangeRatesEndpoint retrieves current exchange rates
func GetExchangeRatesEndpoint(c *gin.Context) {
	rates, err := services.GetExchangeRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Exchange rates retrieved successfully",
		"data":    rates,
	})
}

// CalculateConversionEndpoint calculates conversion amounts without performing conversion
func CalculateConversionEndpoint(c *gin.Context) {
	var req types.ConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	calculation, err := services.CalculateConversion(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Conversion calculated successfully",
		"data":    calculation,
	})
}

// ExecuteConversionEndpoint performs currency conversion
func ExecuteConversionEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req types.ConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	response, err := services.ExecuteConversion(claims.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Currency conversion completed successfully",
		"data":    response,
	})
}

// GetConversionHistoryEndpoint retrieves conversion history for the user
func GetConversionHistoryEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	fromCurrency := c.Query("from_currency")
	toCurrency := c.Query("to_currency")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	pagination := types.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	history, paginationResp, err := services.GetConversionHistory(claims.UserID, pagination, status, fromCurrency, toCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "Conversion history retrieved successfully",
		"data":       history,
		"pagination": paginationResp,
	})
}
