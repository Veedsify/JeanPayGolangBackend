package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

func GetAdminDashboardStatistics(c *gin.Context) {
	dashboard, err := services.GetAdminDashboardStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve dashboard statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Dashboard statistics retrieved successfully",
		"data":    dashboard,
	})
}

func GetAdminUsersAll(c *gin.Context) {
	var params types.UserQueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	users, err := services.GetAdminUsersAll(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Users retrieved successfully",
		"data":    users,
	})
}

func AdminUsersDetails(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}
	user_id, err := libs.ConvertStringToUint(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid User ID format",
			"details": err.Error(),
		})
		return
	}
	userDetails, err := services.GetAdminUserDetails(user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User details retrieved successfully",
		"data":    userDetails,
	})
}

func AdminUserUpdate(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	var request types.UserDetailsEditFields
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userId, err := libs.ConvertStringToUint(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid User ID format",
			"details": err.Error(),
		})
		return
	}

	response, err := services.UpdateAdminUser(userId, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User updated successfully",
		"data":    response,
	})
}

func GetAdminTransactionsAll(c *gin.Context) {
	var params types.AdminTransactionQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	transactions, err := services.GetAdminTransactionsAll(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transactions retrieved successfully",
		"data":    transactions,
	})
}

func GetAdminTransactionDetails(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	transaction, err := services.GetAdminTransactionDetails(transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Transaction not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction details retrieved successfully",
		"data":    transaction,
	})
}

func ApproveAdminTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	// Get user from context (set by middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(*libs.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.ApproveAdminTransaction(transactionID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to approve transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

func RejectAdminTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	var request types.RejectTransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(*libs.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.RejectAdminTransaction(transactionID, request.Reason, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to reject transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

func AdminTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	status, err := services.GetAdminTransactionStatus(transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Transaction not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction status retrieved successfully",
		"data":    status,
	})
}

func AdminTransactionsOverview(c *gin.Context) {
	overview, err := services.GetAdminTransactionsOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve transactions overview",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transactions overview retrieved successfully",
		"data":    overview,
	})
}

func AdminRatesHistory(c *gin.Context) {
	cursor := c.Query("cursor")
	limit := c.Query("limit")
	if limit == "" {
		limit = "10"
	}
	if cursor == "" {
		cursor = "0"
	}
	search := c.Query("search")
	limitInt, err := libs.ConvertStringToInt(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid limit parameter",
			"details": err.Error(),
		})
		return
	}
	cursorUint, err := libs.ConvertStringToUint(cursor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid cursor parameter",
			"details": err.Error(),
		})
		return
	}

	rates, nextCursor, err := services.GetAdminRatesHistory(cursorUint, limitInt, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve rates history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "Rates history retrieved successfully",
		"nextCursor": nextCursor,
		"data":       rates,
	})
}

func AdminRatesAdd(c *gin.Context) {
	var request types.CreateRateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(*libs.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	rate, err := services.AddAdminRate(request, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add rate",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Rate added successfully",
		"data":    rate,
	})
}

// Additional user management controllers
func BlockUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	var request types.BlockUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(*libs.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.BlockUser(userID, request.Reason, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to block user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

func UnblockUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(*libs.JWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.UnblockUser(userID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to unblock user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

// GetUserTransactions retrieves transaction history for a specific user
func GetUserTransactions(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	var params types.AdminTransactionQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	userId, err := libs.ConvertStringToUint(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid User ID format",
			"details": err.Error(),
		})
		return
	}

	transactions, err := services.GetAdminUserTransactionHistory(userId, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve user transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User transactions retrieved successfully",
		"data":    transactions,
	})
}

// GetUserWallet retrieves wallet details for a specific user
func GetUserWallet(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	userId, err := libs.ConvertStringToUint(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid User ID format",
			"details": err.Error(),
		})
		return
	}

	wallet, err := services.GetUserWalletDetails(userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Failed to retrieve user wallet",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User wallet retrieved successfully",
		"data":    wallet,
	})
}

// SearchUsers searches for users based on query string
func SearchUsers(c *gin.Context) {
	var request struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	users, err := services.SearchUsers(request.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to search users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Users search completed successfully",
		"data":    users,
	})
}

// UpdateRate updates an existing exchange rate
func UpdateRate(c *gin.Context) {
	rateIDStr := c.Param("id")
	rateID, err := strconv.ParseUint(rateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid rate ID",
		})
		return
	}

	var request types.UpdateRateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	rate, err := services.UpdateAdminRate(uint(rateID), request, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update rate",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Rate updated successfully",
		"data":    rate,
	})
}

// ToggleRateStatus activates or deactivates a rate
func ToggleRateStatus(c *gin.Context) {
	rateIDStr := c.Param("id")
	rateID, err := strconv.ParseUint(rateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid rate ID",
		})
		return
	}

	var request struct {
		Active bool `json:"active"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.ToggleRateStatus(uint(rateID), request.Active, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update rate status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

// DeleteRate deletes an exchange rate
func DeleteRate(c *gin.Context) {
	rateIDStr := c.Param("id")
	rateID, err := strconv.ParseUint(rateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid rate ID",
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.DeleteRate(uint(rateID), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to delete rate",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

// GetPendingTransactions retrieves all pending transactions
func GetPendingTransactions(c *gin.Context) {
	var params types.AdminTransactionQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	transactions, err := services.GetTransactionsByStatus(models.TransactionPending, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve pending transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Pending transactions retrieved successfully",
		"data":    transactions,
	})
}

// GetFailedTransactions retrieves all failed transactions
func GetFailedTransactions(c *gin.Context) {
	var params types.AdminTransactionQuery
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	transactions, err := services.GetTransactionsByStatus(models.TransactionFailed, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve failed transactions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Failed transactions retrieved successfully",
		"data":    transactions,
	})
}

// AddTransactionNote adds an admin note to a transaction
func AddTransactionNote(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	var request struct {
		Note string `json:"note" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Invalid user context",
		})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin privileges required",
		})
		return
	}

	response, err := services.AddTransactionNote(transactionID, request.Note, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add transaction note",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": response.Message,
		"data":    response,
	})
}

// GetUserActivityLogs retrieves activity logs for a specific user
func GetUserActivityLogs(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	activities, err := services.GetUserActivityLogs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve user activity logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User activity logs retrieved successfully",
		"data":    activities,
	})
}
