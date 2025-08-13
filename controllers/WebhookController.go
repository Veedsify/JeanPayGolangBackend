package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// HandlePaystackWebhookEndpoint processes Paystack webhook events
func HandlePaystackWebhookEndpoint(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Failed to read request body",
		})
		return
	}

	// Get the signature from headers
	signature := c.GetHeader("X-Paystack-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Missing webhook signature",
		})
		return
	}

	// Process the webhook
	err = services.HandlePaystackWebhook(body, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Webhook processed successfully",
	})
}

// HandleMomoWebhookEndpoint processes Mobile Money webhook events
func HandleMomoWebhookEndpoint(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Failed to read request body",
		})
		return
	}

	// Get the signature from headers (adjust header name based on your MoMo provider)
	signature := c.GetHeader("X-Momo-Signature")
	if signature == "" {
		// Try alternative header names
		signature = c.GetHeader("X-Signature")
		if signature == "" {
			signature = c.GetHeader("Authorization")
		}
	}

	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Missing webhook signature",
		})
		return
	}

	// Process the webhook
	err = services.HandleMomoWebhook(body, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Webhook processed successfully",
	})
}

// GetWebhookEventLogsEndpoint retrieves webhook event logs (admin only)
func GetWebhookEventLogsEndpoint(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin access required",
		})
		return
	}

	// Parse query parameters
	provider := c.Query("provider")
	status := c.Query("status")

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page := 1
	limit := 50

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	pagination := types.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	logs, paginationResp, err := services.GetWebhookEventLogs(pagination, provider, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"message":    "Webhook logs retrieved successfully",
		"data":       logs,
		"pagination": paginationResp,
	})
}

// TestWebhookEndpoint allows testing webhook functionality (development only)
func TestWebhookEndpoint(c *gin.Context) {
	// Only allow in development environment
	if !constants.IsDevelopment() {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Test endpoint only available in development",
		})
		return
	}

	var req struct {
		Provider string      `json:"provider" binding:"required"`
		Event    string      `json:"event" binding:"required"`
		Data     interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Create test webhook payload
	testPayload := map[string]interface{}{
		"event": req.Event,
		"data":  req.Data,
	}

	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to create test payload",
		})
		return
	}

	// Generate test signature
	testSignature := "test_signature_" + libs.GenerateRandomString(32)

	// Process based on provider
	var processErr error
	switch req.Provider {
	case "paystack":
		processErr = services.HandlePaystackWebhook(payloadBytes, testSignature)
	case "momo":
		processErr = services.HandleMomoWebhook(payloadBytes, testSignature)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Unsupported provider",
		})
		return
	}

	if processErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": processErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Test webhook processed successfully",
		"data": gin.H{
			"provider": req.Provider,
			"event":    req.Event,
			"payload":  testPayload,
		},
	})
}

// WebhookHealthCheckEndpoint provides health check for webhook endpoints
func WebhookHealthCheckEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Webhook endpoints are healthy",
		"data": gin.H{
			"status":    "ok",
			"timestamp": time.Now(),
			"endpoints": []string{
				"/webhooks/paystack",
				"/webhooks/momo",
			},
		},
	})
}
