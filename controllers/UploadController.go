package controllers

import (
	"net/http"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/gin-gonic/gin"
)

// UploadReceiptEndpoint handles receipt file uploads for transactions
func UploadReceiptEndpoint(c *gin.Context) {
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

	// Get transaction ID from form data
	transactionID := c.PostForm("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	// Verify that the transaction belongs to the user
	var transaction models.Transaction
	if err := database.DB.Where("transaction_id = ? AND user_id = ?", transactionID, userID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Transaction not found or access denied",
		})
		return
	}

	// Get the uploaded file
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "No file uploaded or invalid file",
		})
		return
	}
	defer file.Close()

	// Initialize Cloudinary service
	cloudinaryService, err := services.NewCloudinaryService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to initialize upload service",
		})
		return
	}

	// Upload file to Cloudinary
	cloudinaryURL, err := cloudinaryService.UploadReceipt(file, header, transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	// Create receipt record in database
	receipt := models.Receipt{
		TransactionID: transactionID,
		CloudinaryURL: cloudinaryURL,
		OriginalName:  header.Filename,
		FileSize:      header.Size,
		FileType:      header.Header.Get("Content-Type"),
		UploadedAt:    time.Now(),
	}

	if err := database.DB.Create(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to save receipt record",
		})
		return
	}

	// Update transaction with receipt URL
	if err := database.DB.Model(&transaction).Update("receipt_url", cloudinaryURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to update transaction with receipt URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Receipt uploaded successfully",
		"data": gin.H{
			"receipt_id":     receipt.ID,
			"cloudinary_url": cloudinaryURL,
			"transaction_id": transactionID,
			"uploaded_at":    receipt.UploadedAt,
		},
	})
}

// GetReceiptEndpoint retrieves receipt information for a transaction
func GetReceiptEndpoint(c *gin.Context) {
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

	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Transaction ID is required",
		})
		return
	}

	// Verify that the transaction belongs to the user (or user is admin)
	var transaction models.Transaction
	query := database.DB.Where("transaction_id = ?", transactionID)
	if !claims.(*libs.JWTClaims).IsAdmin {
		query = query.Where("user_id = ?", userID)
	}
	
	if err := query.First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Transaction not found or access denied",
		})
		return
	}

	// Get receipt for the transaction
	var receipt models.Receipt
	if err := database.DB.Where("transaction_id = ?", transactionID).First(&receipt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "No receipt found for this transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  receipt,
	})
}