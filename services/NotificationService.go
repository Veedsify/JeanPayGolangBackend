package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/types"
)

// NotificationData represents notification information
type NotificationData struct {
	ID        uint                    `json:"id"`
	UserID    uint                    `json:"userId"`
	Type      models.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Read      bool                    `json:"read"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

// NotificationSummary represents notification summary
type NotificationSummary struct {
	Notifications []NotificationData `json:"notifications"`
	UnreadCount   int64              `json:"unreadCount"`
}

// GetAllNotifications retrieves all notifications for a user
func GetAllNotifications(userID uint32, pagination types.PaginationRequest) (*NotificationSummary, *types.PaginationResponse, error) {
	if userID == 0 {
		return nil, nil, errors.New("user ID is required")
	}

	// Build query
	query := database.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Count unread notifications
	var unreadCount int64
	database.DB.Model(&models.Notification{}).Where("user_id = ? AND read = ?", userID, false).Count(&unreadCount)

	// Calculate pagination
	page, limit := pagination.GetValidatedParams()
	offset := (page - 1) * limit

	// Find notifications
	var notifications []models.Notification
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find notifications: %w", err)
	}

	// Convert to response format
	var notificationData []NotificationData
	for _, notification := range notifications {
		notificationData = append(notificationData, NotificationData{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Type:      models.NotificationType(notification.Type),
			Title:     getNotificationTitle(notification.Title),
			Message:   notification.Message,
			Read:      notification.Read,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.UpdatedAt,
		})
	}

	summary := &NotificationSummary{
		Notifications: notificationData,
		UnreadCount:   unreadCount,
	}

	// Create pagination response
	paginationResp := types.NewPaginationResponse(page, limit, total)

	return summary, paginationResp, nil
}

// MarkNotificationRead marks a specific notification as read
func MarkNotificationRead(userID uint32, notificationID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if notificationID == 0 {
		return errors.New("notification ID is required")
	}

	// Find and update notification
	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

// MarkAllNotificationsRead marks all notifications as read for a user
func MarkAllNotificationsRead(userID uint32) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	// Update all unread notifications for the user
	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// CreateNotification creates a new notification for a user
func CreateNotification(userID uint, notificationType, message string) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if notificationType == "" {
		return errors.New("notification type is required")
	}

	if message == "" {
		return errors.New("message is required")
	}

	notification := models.Notification{
		UserID:  userID,
		Type:    models.NotificationType(notificationType),
		Message: message,
		Read:    false,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// CreateTransactionNotification creates a notification for transaction events
func CreateTransactionNotification(userID uint, transactionType, amount, currency, transactionID string) error {
	message := fmt.Sprintf("Your %s of %s %s has been processed. Transaction ID: %s",
		transactionType, currency, amount, transactionID)

	notificationType := "transaction_" + transactionType
	return CreateNotification(userID, notificationType, message)
}

// CreateWelcomeNotification creates a welcome notification for new users
func CreateWelcomeNotification(userID uint, firstName string) error {
	message := fmt.Sprintf("Welcome to JeanPay, %s! Your account has been successfully created.", firstName)
	return CreateNotification(userID, "welcome", message)
}

// CreateSecurityNotification creates a security-related notification
func CreateSecurityNotification(userID uint, action string) error {
	message := fmt.Sprintf("Security alert: %s on your account", action)
	return CreateNotification(userID, "security", message)
}

// CreateSystemNotification creates a system notification
func CreateSystemNotification(userID uint, message string) error {
	return CreateNotification(userID, "system", message)
}

// DeleteNotification deletes a notification
func DeleteNotification(userID uint, notificationID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if notificationID == 0 {
		return errors.New("notification ID is required")
	}

	result := database.DB.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

// DeleteAllNotifications deletes all notifications for a user
func DeleteAllNotifications(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if err := database.DB.Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("failed to delete notifications: %w", err)
	}

	return nil
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func GetUnreadNotificationCount(userID uint) (int64, error) {
	if userID == 0 {
		return 0, errors.New("user ID is required")
	}

	var count int64
	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// GetNotificationsByType retrieves notifications by type for a user
func GetNotificationsByType(userID uint, notificationType string, pagination types.PaginationRequest) ([]NotificationData, *types.PaginationResponse, error) {
	if userID == 0 {
		return nil, nil, errors.New("user ID is required")
	}

	if notificationType == "" {
		return nil, nil, errors.New("notification type is required")
	}

	// Build query
	query := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND type = ?", userID, notificationType)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Calculate pagination
	page, limit := pagination.GetValidatedParams()
	offset := (page - 1) * limit

	// Find notifications
	var notifications []models.Notification
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find notifications: %w", err)
	}

	// Convert to response format
	var notificationData []NotificationData
	for _, notification := range notifications {
		notificationData = append(notificationData, NotificationData{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Type:      notification.Type,
			Title:     getNotificationTitle(notification.Title),
			Message:   notification.Message,
			Read:      notification.Read,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.UpdatedAt,
		})
	}

	// Create pagination response
	paginationResp := types.NewPaginationResponse(page, limit, total)

	return notificationData, paginationResp, nil
}

// BulkCreateNotifications creates multiple notifications at once
func BulkCreateNotifications(notifications []models.Notification) error {
	if len(notifications) == 0 {
		return errors.New("no notifications to create")
	}

	if err := database.DB.Create(&notifications).Error; err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	return nil
}

// SendNotificationToAllUsers creates a notification for all active users
func SendNotificationToAllUsers(notificationType, message string) error {
	if notificationType == "" {
		return errors.New("notification type is required")
	}

	if message == "" {
		return errors.New("message is required")
	}

	// Get all active users
	var users []models.User
	if err := database.DB.Select("user_id").Where("is_verified = ? AND is_blocked = ?", true, false).Find(&users).Error; err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	// Prepare notifications
	var notifications []models.Notification
	for _, user := range users {
		notifications = append(notifications, models.Notification{
			UserID:  user.ID,
			Type:    models.NotificationType(notificationType),
			Message: message,
			Read:    false,
		})
	}

	// Bulk create notifications
	return BulkCreateNotifications(notifications)
}

// Helper functions

// getNotificationTitle returns a user-friendly title for notification types
func getNotificationTitle(notificationType string) string {
	switch notificationType {
	case "welcome":
		return "Welcome to JeanPay!"
	case "transaction_deposit":
		return "Deposit Successful"
	case "transaction_withdrawal":
		return "Withdrawal Processed"
	case "transaction_conversion":
		return "Currency Conversion"
	case "transaction_transfer":
		return "Transfer Complete"
	case "security":
		return "Security Alert"
	case "system":
		return "System Notification"
	case "promotion":
		return "Special Offer"
	case "maintenance":
		return "Maintenance Notice"
	default:
		return "Notification"
	}
}

// GetAllNotificationsService retrieves all notifications for a user with query parameters
func GetAllNotificationsService(userID string, query types.NotificationQuery) (*types.GetNotificationsResponse, error) {
	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	// Build GORM query
	db := database.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	// Apply filters
	switch query.ReadStatus {
	case "read":
		db = db.Where("read = ?", true)
	case "unread":
		db = db.Where("read = ?", false)
	}

	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get notifications
	var notifications []models.Notification
	offset := (query.Page - 1) * query.Limit
	err := db.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&notifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Get unread count
	var unreadCount int64
	database.DB.Model(&models.Notification{}).Where("user_id = ? AND read = ?", userID, false).Count(&unreadCount)

	// Convert to response format
	response := &types.GetNotificationsResponse{
		Notifications: types.ToNotificationsResponse(notifications),
		Total:         total,
		Page:          query.Page,
		Limit:         query.Limit,
		TotalPages:    int((total + int64(query.Limit) - 1) / int64(query.Limit)),
		UnreadCount:   unreadCount,
	}

	return response, nil
}

// MarkNotificationReadService marks a notification as read
func MarkNotificationReadService(notificationID string, userID string) error {
	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found or already read")
	}

	return nil
}

// MarkAllNotificationsReadService marks all notifications as read for a user
func MarkAllNotificationsReadService(userID string) (int64, error) {
	result := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetUnreadNotificationCountService returns unread notification count for a user
func GetUnreadNotificationCountService(userID string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// CreateNotificationService creates a new notification
func CreateNotificationService(request types.CreateNotificationRequest, adminID string) (*types.CreateNotificationResponse, error) {
	// Convert string UserID to uint32
	notification := models.Notification{
		UserID:  request.UserID,
		Type:    request.Type,
		Message: request.Message,
		Read:    false,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	response := &types.CreateNotificationResponse{
		ID:        uint32(notification.ID),
		UserID:    strconv.Itoa(int(notification.UserID)),
		Type:      models.NotificationType(notification.Type),
		Title:     request.Title,
		Message:   notification.Message,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
	}

	return response, nil
}

// DeleteNotificationService deletes a notification
func DeleteNotificationService(notificationID string, userID string, isAdmin bool) error {
	query := database.DB.Where("id = ?", notificationID)

	// If not admin, restrict to user's own notifications
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	result := query.Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

// GetNotificationDetailsService returns detailed notification information
func GetNotificationDetailsService(notificationID string, userID string, isAdmin bool) (*types.NotificationDetailsResponse, error) {
	var notification models.Notification
	query := database.DB.Where("id = ?", notificationID)

	// If not admin, restrict to user's own notifications
	if !isAdmin {
		query = query.Where("user_id = ?", userID)
	}

	err := query.First(&notification).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	response := &types.NotificationDetailsResponse{
		ID:        uint32(notification.ID),
		UserID:    strconv.Itoa(int(notification.UserID)),
		Type:      notification.Type,
		Title:     getNotificationTitle(notification.Title),
		Message:   notification.Message,
		Data:      "",
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}

	return response, nil
}

// GetRecentNotifications retrieves recent notifications for a user
func GetRecentNotifications(userID uint32, limit int) ([]types.NotificationResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	var notifications []models.Notification
	err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent notifications: %w", err)
	}

	// Convert to response format
	var notificationResponses []types.NotificationResponse
	for _, notification := range notifications {
		notificationResponses = append(notificationResponses, types.NotificationResponse{
			UserID:    notification.UserID,
			Type:      notification.Type,
			Message:   notification.Message,
			Read:      notification.Read,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.UpdatedAt,
		})
	}

	return notificationResponses, nil
}
