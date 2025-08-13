package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// AdminLoginEndpoint handles admin login requests
func AdminLoginEndpoint(c *gin.Context) {
	var loginRequest types.AdminLoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	response, err := services.AdminLogin(loginRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	// Set cookies for authentication
	c.SetCookie("admin_token", response.Token, 3600*24, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Admin login successful",
		"data":    response,
	})
}

// GetAdminDashboardEndpoint returns admin dashboard overview
func GetAdminDashboardStatistics(c *gin.Context) {
	stats, err := services.GetAdminDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  stats,
	})
}





// GetAllTransactionsAdminEndpoint returns paginated transactions for admin
func GetAllTransactionsAdminEndpoint(c *gin.Context) {
	var query types.TransactionQueryParams

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
	query.Search = c.Query("search")

	response, err := services.GetAllTransactionsAdminService(query)
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

// GetTransactionDetailsAdminEndpoint returns detailed transaction information
func GetTransactionDetailsAdminEndpoint(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	response, err := services.GetTransactionDetailsAdminService(transactionID)
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

// ApproveTransactionEndpoint approves a pending transaction
func ApproveTransactionEndpoint(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	// Get admin ID from JWT token (assumes middleware sets this)
	adminIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	adminID := adminIDInterface.(uint32)

	err := services.ApproveTransactionService(transactionID, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction approved successfully",
		"data": gin.H{
			"transactionId": transactionID,
			"status":        "approved",
		},
	})
}

// RejectTransactionEndpoint rejects a pending transaction
func RejectTransactionEndpoint(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	// Get admin ID from JWT token (assumes middleware sets this)
	adminIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	adminID := (adminIDInterface.(uint32))

	err := services.RejectTransactionService(transactionID, adminID, request.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction rejected successfully",
		"data": gin.H{
			"transactionId": transactionID,
			"status":        "rejected",
			"reason":        request.Reason,
		},
	})
}

// GetAllUsersAdminEndpoint returns paginated users for admin
func GetAllUsersAdminEndpoint(c *gin.Context) {
	var query types.UserQueryParams

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
	query.Blocked = c.Query("blocked")
	query.Search = c.Query("search")

	response, err := services.GetAllUsersAdminService(query)
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

// BlockUserEndpoint blocks a user account
func BlockUserEndpoint(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	// Get admin ID from JWT token (assumes middleware sets this)
	adminIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	adminID := (adminIDInterface.(uint32))

	err := services.BlockUserService(userID, adminID, request.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User blocked successfully",
		"data": gin.H{
			"userId":  userID,
			"blocked": true,
			"reason":  request.Reason,
		},
	})
}

// UnblockUserEndpoint unblocks a user account
func UnblockUserEndpoint(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "User ID is required",
		})
		return
	}

	// Get admin ID from JWT token (assumes middleware sets this)
	adminIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	adminID := (adminIDInterface.(uint32))

	err := services.UnblockUserService(userID, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User unblocked successfully",
		"data": gin.H{
			"userId":  userID,
			"blocked": false,
		},
	})
}
