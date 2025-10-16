package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// GetWalletBalanceEndpoint retrieves wallet balance for authenticated user
func GetWalletBalanceEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found in context", "error": true})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)
	balance, err := services.GetWalletBalance(claims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Wallet balance retrieved successfully",
		"data":    balance,
	})
}

// TopUpWalletEndpoint initiates wallet top-up
func TopUpWalletEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userID := claims.(*libs.JWTClaims).ID

	var req types.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	response, err := services.TopUpWallet(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Wallet top-up initiated successfully",
		"data":    response,
	})
}

// WithdrawFromWalletEndpoint initiates wallet withdrawal
func WithdrawFromWalletEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userId := claims.(*libs.JWTClaims).ID

	var req types.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	response, err := services.WithdrawFromWallet(userId, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Wallet withdrawal initiated successfully",
		"data":    response,
	})
}

// GetWalletHistoryEndpoint retrieves wallet transaction history
func GetWalletHistoryEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	txType := c.Query("type")
	status := c.Query("status")

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

	history, paginationResp, err := services.GetWalletHistory(userID.(uint32), pagination, txType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "Wallet history retrieved successfully",
		"data":       history,
		"pagination": paginationResp,
	})
}

// UpdateWalletAfterPaymentEndpoint updates wallet after successful payment (internal use)
func UpdateWalletAfterPaymentEndpoint(c *gin.Context) {
	var req struct {
		UserID   uint    `json:"userId" binding:"required"`
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := services.UpdateWalletAfterPayment(req.UserID, req.Currency, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Wallet updated successfully",
	})
}

// GetTopUpDetailsEndpoint retrieves topup transaction details
func GetTopUpDetailsEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userID := claims.(*libs.JWTClaims).ID
	transactionID := c.Param("id")

	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	topupDetails, err := services.GetTopUpDetails(userID, transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Topup details retrieved successfully",
		"data":    topupDetails,
	})
}

// CalculateWithdrawalFeeEndpoint calculates withdrawal fee for given amount and currency
func CalculateWithdrawalFeeEndpoint(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required,oneof=NGN GHS"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	feeCalculation := services.CalculateWithdrawalFee(req.Amount, req.Currency)

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal fee calculated successfully",
		"data":    feeCalculation,
	})
}

// ValidateWithdrawalEndpoint validates withdrawal request before processing
func ValidateWithdrawalEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userID := claims.(*libs.JWTClaims).ID

	var req types.WithdrawalValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Override userID from token
	req.UserID = userID

	// Get user's wallet balance
	wallets, err := services.GetWalletBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to get wallet balance",
		})
		return
	}

	// Find balance for requested currency
	var availableBalance float64
	for _, wallet := range wallets {
		if wallet.Currency == req.Currency {
			availableBalance = wallet.Balance
			break
		}
	}

	// Calculate fee
	feeCalc := services.CalculateWithdrawalFee(req.Amount, req.Currency)

	// Validate withdrawal
	err = services.ValidateWithdrawalLimits(userID, req.Amount, req.Currency)

	response := types.WithdrawalValidationResponse{
		IsValid:          err == nil,
		AvailableBalance: availableBalance,
		FeeCalculation:   feeCalc,
	}

	if err != nil {
		response.ErrorMessage = err.Error()
	}

	statusCode := http.StatusOK
	if !response.IsValid {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"error": !response.IsValid,
		"message": func() string {
			if response.IsValid {
				return "Withdrawal validation successful"
			}
			return "Withdrawal validation failed"
		}(),
		"data": response,
	})
}

// GetWithdrawalDetailsEndpoint retrieves withdrawal transaction details
func GetWithdrawalDetailsEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userID := claims.(*libs.JWTClaims).ID
	transactionID := c.Param("id")

	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	withdrawalDetails, err := services.GetWithdrawalDetails(userID, transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawal details retrieved successfully",
		"data":    withdrawalDetails,
	})
}

// GetUserWithdrawalsEndpoint retrieves user's withdrawal history
func GetUserWithdrawalsEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userID := claims.(*libs.JWTClaims).ID

	var req types.GetWithdrawalsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	withdrawals, err := services.GetUserWithdrawals(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Withdrawals retrieved successfully",
		"data":    withdrawals,
	})
}
