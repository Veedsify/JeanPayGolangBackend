package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// GetAllWithdrawals retrieves all withdrawal requests with pagination and filters
func GetAllWithdrawals(c *gin.Context) {
	var params types.WithdrawalQueryParams

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}

	params.Status = c.Query("status")
	params.Currency = c.Query("currency")
	params.Method = c.Query("method")
	params.Search = c.Query("search")
	params.FromDate = c.Query("fromDate")
	params.ToDate = c.Query("toDate")

	if minAmount := c.Query("minAmount"); minAmount != "" {
		if min, err := strconv.ParseFloat(minAmount, 64); err == nil {
			params.MinAmount = &min
		}
	}
	if maxAmount := c.Query("maxAmount"); maxAmount != "" {
		if max, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			params.MaxAmount = &max
		}
	}

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	withdrawals, pagination, err := services.GetAllWithdrawals(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve withdrawals: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawals retrieved successfully",
		"data": gin.H{
			"withdrawals": withdrawals,
			"pagination":  pagination,
		},
	})
}

// GetWithdrawalStats retrieves withdrawal statistics
func GetWithdrawalStats(c *gin.Context) {
	stats, err := services.GetWithdrawalStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve withdrawal statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal statistics retrieved successfully",
		"data":    stats,
	})
}

// GetWithdrawalDetails retrieves detailed information for a specific withdrawal
func GetWithdrawalDetails(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	withdrawal, err := services.GetAdminWithdrawalDetails(withdrawalId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Withdrawal not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal details retrieved successfully",
		"data":    withdrawal,
	})
}

// ApproveWithdrawal approves a pending withdrawal request
func ApproveWithdrawal(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var request types.ApproveWithdrawalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	result, err := services.ApproveWithdrawal(withdrawalId, request.AdminNotes, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to approve withdrawal: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal approved successfully",
		"data":    result,
	})
}

// RejectWithdrawal rejects a pending withdrawal request
func RejectWithdrawal(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var request types.RejectWithdrawalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	if request.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Rejection reason is required",
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	result, err := services.RejectWithdrawal(withdrawalId, request.Reason, request.AdminNotes, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to reject withdrawal: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal rejected successfully",
		"data":    result,
	})
}

// GetPendingWithdrawals retrieves all pending withdrawal requests
func GetPendingWithdrawals(c *gin.Context) {
	var params types.WithdrawalQueryParams

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}

	params.Currency = c.Query("currency")
	params.Method = c.Query("method")
	params.Search = c.Query("search")
	params.Status = "PENDING" // Force pending status

	// Set defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	withdrawals, pagination, err := services.GetAllWithdrawals(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve pending withdrawals: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Pending withdrawals retrieved successfully",
		"data": gin.H{
			"withdrawals": withdrawals,
			"pagination":  pagination,
		},
	})
}

// UpdateWithdrawalStatus updates the status of a withdrawal request
func UpdateWithdrawalStatus(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var request types.UpdateWithdrawalStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	result, err := services.UpdateWithdrawalStatus(withdrawalId, request.Status, request.Reason, request.AdminNotes, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update withdrawal status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal status updated successfully",
		"data":    result,
	})
}

// AddWithdrawalNote adds an admin note to a withdrawal request
func AddWithdrawalNote(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	var request struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Note is required",
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Admin authentication required",
		})
		return
	}

	result, err := services.AddWithdrawalNote(withdrawalId, request.Note, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add withdrawal note: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal note added successfully",
		"data":    result,
	})
}

// GetWithdrawalNotes retrieves all notes for a withdrawal request
func GetWithdrawalNotes(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	notes, err := services.GetWithdrawalNotes(withdrawalId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve withdrawal notes: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal notes retrieved successfully",
		"data":    notes,
	})
}

// GetWithdrawalHistory retrieves status history for a withdrawal request
func GetWithdrawalHistory(c *gin.Context) {
	withdrawalId := c.Param("id")
	if withdrawalId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Withdrawal ID is required",
		})
		return
	}

	history, err := services.GetWithdrawalHistory(withdrawalId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve withdrawal history: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal history retrieved successfully",
		"data":    history,
	})
}
