package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// GetPaymentAccounts retrieves all payment accounts with optional filters
func GetPaymentAccounts(c *gin.Context) {
	var params struct {
		Currency    string `json:"currency"`
		AccountType string `json:"account_type"`
		IsActive    *bool  `json:"is_active"`
		Search      string `json:"search"`
		Page        int    `json:"page"`
		Limit       int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}

	accounts, total, err := services.GetPaymentAccounts(params.Currency, params.AccountType, params.IsActive, params.Search, params.Page, params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve payment accounts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payment accounts retrieved successfully",
		"data":    accounts,
		"pagination": gin.H{
			"total":       total,
			"page":        params.Page,
			"limit":       params.Limit,
			"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
		},
	})
}

// CreatePaymentAccount creates a new payment account
func CreatePaymentAccount(c *gin.Context) {
	var params types.CreatePaymentAccountRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin ID not found in context",
		})
		return
	}

	adminID := claims.(*libs.JWTClaims).ID

	account, err := services.CreatePaymentAccount(params, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to create payment account",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Payment account created successfully",
		"data":    account,
	})
}

// UpdatePaymentAccount updates an existing payment account
func UpdatePaymentAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Account ID is required",
		})
		return
	}

	id, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid account ID",
		})
		return
	}

	var params types.UpdatePaymentAccountRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	account, err := services.UpdatePaymentAccount(uint(id), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update payment account",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payment account updated successfully",
		"data":    account,
	})
}

// DeletePaymentAccount deletes a payment account
func DeletePaymentAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Account ID is required",
		})
		return
	}

	id, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid account ID",
		})
		return
	}

	err = services.DeletePaymentAccount(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete payment account",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payment account deleted successfully",
	})
}

// TogglePaymentAccountStatus toggles the active status of a payment account
func TogglePaymentAccountStatus(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Account ID is required",
		})
		return
	}

	id, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid account ID",
		})
		return
	}

	account, err := services.TogglePaymentAccountStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to toggle payment account status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payment account status updated successfully",
		"data":    account,
	})
}

// GetPaymentAccountByID retrieves a single payment account by ID
func GetPaymentAccountByID(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Account ID is required",
		})
		return
	}

	id, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid account ID",
		})
		return
	}

	account, err := services.GetPaymentAccountByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve payment account",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Payment account retrieved successfully",
		"data":    account,
	})
}

// GetActivePaymentAccounts retrieves active payment accounts for topup functionality
func GetActivePaymentAccounts(c *gin.Context) {
	currency := c.Query("currency")
	accountType := c.Query("account_type")

	accounts, err := services.GetActivePaymentAccounts(currency, accountType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve active payment accounts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Active payment accounts retrieved successfully",
		"data":    accounts,
	})
}
