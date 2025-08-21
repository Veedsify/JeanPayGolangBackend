package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/gin-gonic/gin"
)

// Create New Transactions
func CreateTransactionEndpoint(c *gin.Context) {
	// Get user ID from JWT token
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID
	var transaction types.NewTransactionRequest
	c.ShouldBindJSON(&transaction)
	if transaction.FromAmount == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "FromAmount is required",
		})
		return
	}
	if transaction.ToAmount == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ToAmount is required",
		})
		return
	}
	if transaction.FromCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "FromCurrency is required",
		})
		return
	}
	if transaction.ToCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ToCurrency is required",
		})
		return
	}
	if transaction.FromCurrency == transaction.ToCurrency {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "FromCurrency and ToCurrency cannot be the same",
		})
		return
	}
	if transaction.MethodOfPayment == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "MethodOfPayment is required",
		})
		return
	}
	response, code, err := services.CreateTransaction(userID, transaction)
	codeError := utils.GetErrorFromCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
			"code":    codeError,
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"error": false,
		"data":  response,
	})
}

// GetUserTransactionHistoryEndpoint returns user's transaction history
func GetUserTransactionHistoryEndpoint(c *gin.Context) {
	// Get user ID from JWT token
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID
	var query types.UserTransactionQuery
	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query.Page = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			query.Limit = l
		}
	}
	query.Status = c.Query("status")
	query.Type = c.Query("type")
	query.Currency = c.Query("currency")
	query.FromDate = c.Query("from_date")
	query.ToDate = c.Query("to_date")
	query.AccountType = c.Query("account_type")
	query.Search = c.Query("search")
	response, err := services.GetUserTransactionHistoryService(userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"data":       response.Transactions,
		"pagination": response.Pagination,
	})
}

// GetTransactionDetailsEndpoint returns transaction details
func GetTransactionDetailsEndpoint(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}
	// Get user ID from JWT token
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	user := claims.(*libs.JWTClaims)
	response, err := services.GetTransactionDetailsService(transactionID, user.ID, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  response,
	})
}

// UpdateTransactionStatusEndpoint updates transaction status (Admin only)
func UpdateTransactionStatusEndpoint(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin access required",
		})
		return
	}
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}
	var request types.UpdateTransactionStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}
	// Get admin ID from JWT token
	adminIDInterface, _ := c.Get("user_id")
	adminID := (adminIDInterface.(uint32))
	response, err := services.UpdateTransactionStatusService(transactionID, request, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction status updated successfully",
		"data":    response,
	})
}

// CreateDepositTransactionEndpoint creates a new deposit transaction
func CreateDepositTransactionEndpoint(c *gin.Context) {
	// Get user ID from JWT token
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID
	var request types.CreateDepositRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}
	response, err := services.CreateDepositTransactionService(userID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Deposit transaction created successfully",
		"data":    response,
	})
}

// CreateWithdrawalTransactionEndpoint creates a new withdrawal transaction
func CreateWithdrawalTransactionEndpoint(c *gin.Context) {
	// Get user ID from JWT token
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID
	var request types.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}
	response, err := services.CreateWithdrawalTransactionService(userID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Withdrawal transaction created successfully",
		"data":    response,
	})
}

// GetUserTransactionStatsEndpoint returns transaction statistics for user
func GetUserTransactionStatsEndpoint(c *gin.Context) {
	// Get user ID from JWT token
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User authentication required",
		})
		return
	}
	userID := strconv.Itoa(int(userIDInterface.(uint32)))
	response, err := services.GetTransactionStatsService(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  response,
	})
}

// FilterTransactionsEndpoint filters transactions based on criteria
func FilterTransactionsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	claims := claimsAny.(*libs.JWTClaims)
	var filterReq types.TransactionFilterRequest
	if err := c.ShouldBindJSON(&filterReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid filter criteria",
			"details": err.Error(),
		})
		return
	}
	response, err := services.FilterTransactions(claims.UserID, filterReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  response,
	})
}
