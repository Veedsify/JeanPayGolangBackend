package controllers

import (
	"net/http"
	"strconv"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

// GetAllNotificationsEndpoint returns all notifications for user
func GetAllNotificationsEndpoint(c *gin.Context) {
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

	var query types.NotificationQuery

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

	query.ReadStatus = c.Query("read_status")
	query.Type = c.Query("type")

	response, err := services.GetAllNotificationsService(userID, query)
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

// MarkNotificationReadEndpoint marks a notification as read
func MarkNotificationReadEndpoint(c *gin.Context) {
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

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Notification ID is required",
		})
		return
	}

	err := services.MarkNotificationReadService(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Notification marked as read",
		"data": gin.H{
			"notification_id": notificationID,
			"read":            true,
		},
	})
}

// MarkAllNotificationsReadEndpoint marks all notifications as read for user
func MarkAllNotificationsReadEndpoint(c *gin.Context) {
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

	count, err := services.MarkAllNotificationsReadService(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "All notifications marked as read",
		"data": gin.H{
			"marked_count": count,
		},
	})
}

// GetUnreadNotificationCountEndpoint returns unread notification count
func GetUnreadNotificationCountEndpoint(c *gin.Context) {
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

	count, err := services.GetUnreadNotificationCountService(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data": gin.H{
			"unread_count": count,
		},
	})
}

// CreateNotificationEndpoint creates a new notification (Admin only)
func CreateNotificationEndpoint(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Admin access required",
		})
		return
	}

	var request types.CreateNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	// Get admin ID from JWT token
	adminIDInterface, _ := c.Get("user_id")
	adminID := strconv.Itoa(int(adminIDInterface.(uint32)))

	response, err := services.CreateNotificationService(request, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Notification created successfully",
		"data":    response,
	})
}

// DeleteNotificationEndpoint deletes a notification
func DeleteNotificationEndpoint(c *gin.Context) {
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

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Notification ID is required",
		})
		return
	}

	// Check if user is admin
	isAdmin, _ := c.Get("is_admin")
	isAdminBool := isAdmin != nil && isAdmin.(bool)

	err := services.DeleteNotificationService(notificationID, userID, isAdminBool)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Notification deleted successfully",
	})
}

// GetRecentNotificationsEndpoint returns recent notifications for user
func GetRecentNotificationsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	notifications, err := services.GetRecentNotifications(claims.UserID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Recent notifications retrieved successfully",
		"data":    notifications,
	})
}

// GetNotificationDetailsEndpoint returns notification details
func GetNotificationDetailsEndpoint(c *gin.Context) {
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

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Notification ID is required",
		})
		return
	}

	// Check if user is admin
	isAdmin, _ := c.Get("is_admin")
	isAdminBool := isAdmin != nil && isAdmin.(bool)

	response, err := services.GetNotificationDetailsService(notificationID, userID, isAdminBool)
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
