package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/hibiken/asynq"
)

const (
	TypeNotificationCreate      = "notification:create"
	TypeNotificationUpdate      = "notification:update"
	TypeNotificationMarkRead    = "notification:mark_read"
	TypeNotificationMarkAllRead = "notification:mark_all_read"
	TypeNotificationDelete      = "notification:delete"
)

// NotificationJobPayload represents the payload for creating a notification
type NotificationJobPayload struct {
	UserID  uint                    `json:"user_id"`
	Type    models.NotificationType `json:"type"`
	Title   string                  `json:"title"`
	Message string                  `json:"message"`
}

// NotificationUpdatePayload represents the payload for updating a notification
type NotificationUpdatePayload struct {
	NotificationID uint   `json:"notification_id"`
	Type           string `json:"type,omitempty"`
	Title          string `json:"title"`
	Message        string `json:"message,omitempty"`
	Read           *bool  `json:"read,omitempty"`
}

// NotificationMarkReadPayload represents the payload for marking a notification as read
type NotificationMarkReadPayload struct {
	NotificationID uint `json:"notification_id"`
	UserID         uint `json:"user_id"`
}

// NotificationMarkAllReadPayload represents the payload for marking all notifications as read for a user
type NotificationMarkAllReadPayload struct {
	UserID uint `json:"user_id"`
}

// NotificationDeletePayload represents the payload for deleting a notification
type NotificationDeletePayload struct {
	NotificationID uint `json:"notification_id"`
	UserID         uint `json:"user_id"`
}

// NotificationJobClient handles notification job creation and queuing
type NotificationJobClient struct {
	client *asynq.Client
}

// NewNotificationJobClient creates a new notification job client
func NewNotificationJobClient() *NotificationJobClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &NotificationJobClient{
		client: client,
	}
}

// Close closes the notification job client
func (njc *NotificationJobClient) Close() error {
	return njc.client.Close()
}

// EnqueueCreateNotification queues a notification creation job
func (njc *NotificationJobClient) EnqueueCreateNotification(userID uint, notificationType models.NotificationType, title string, message string) error {
	payload := NotificationJobPayload{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationCreate, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := njc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue notification creation task: %w", err)
	}

	log.Printf("Enqueued notification creation task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueUpdateNotification queues a notification update job
func (njc *NotificationJobClient) EnqueueUpdateNotification(notificationID uint, notificationType, message string, read *bool) error {
	payload := NotificationUpdatePayload{
		NotificationID: notificationID,
		Type:           notificationType,
		Message:        message,
		Read:           read,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal notification update payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationUpdate, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(3 * time.Minute),
	}

	info, err := njc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue notification update task: %w", err)
	}

	log.Printf("Enqueued notification update task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueMarkNotificationRead queues a job to mark a notification as read
func (njc *NotificationJobClient) EnqueueMarkNotificationRead(notificationID, userID uint) error {
	payload := NotificationMarkReadPayload{
		NotificationID: notificationID,
		UserID:         userID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal mark read payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationMarkRead, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("low"),
		asynq.MaxRetry(2),
		asynq.Timeout(2 * time.Minute),
	}

	info, err := njc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue mark notification read task: %w", err)
	}

	log.Printf("Enqueued mark notification read task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueMarkAllNotificationsRead queues a job to mark all notifications as read for a user
func (njc *NotificationJobClient) EnqueueMarkAllNotificationsRead(userID uint) error {
	payload := NotificationMarkAllReadPayload{
		UserID: userID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal mark all read payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationMarkAllRead, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("low"),
		asynq.MaxRetry(2),
		asynq.Timeout(3 * time.Minute),
	}

	info, err := njc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue mark all notifications read task: %w", err)
	}

	log.Printf("Enqueued mark all notifications read task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueDeleteNotification queues a notification deletion job
func (njc *NotificationJobClient) EnqueueDeleteNotification(notificationID, userID uint) error {
	payload := NotificationDeletePayload{
		NotificationID: notificationID,
		UserID:         userID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete notification payload: %w", err)
	}

	task := asynq.NewTask(TypeNotificationDelete, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("low"),
		asynq.MaxRetry(2),
		asynq.Timeout(2 * time.Minute),
	}

	info, err := njc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue delete notification task: %w", err)
	}

	log.Printf("Enqueued delete notification task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// Worker functions

// HandleCreateNotificationTask handles notification creation
func HandleCreateNotificationTask(ctx context.Context, t *asynq.Task) error {
	var payload NotificationJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal notification payload: %v: %w", err, asynq.SkipRetry)
	}

	// Validate payload
	if payload.UserID == 0 {
		return fmt.Errorf("user_id is required: %w", asynq.SkipRetry)
	}
	if payload.Type == "" {
		return fmt.Errorf("notification type is required: %w", asynq.SkipRetry)
	}
	if payload.Message == "" {
		return fmt.Errorf("notification message is required: %w", asynq.SkipRetry)
	}

	notification := models.Notification{
		UserID:  payload.UserID,
		Title:   payload.Title,
		Type:    payload.Type,
		Message: payload.Message,
		Read:    false,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("Notification created successfully for user_id: %d, type: %s", payload.UserID, payload.Type)
	return nil
}

// HandleUpdateNotificationTask handles notification updates
func HandleUpdateNotificationTask(ctx context.Context, t *asynq.Task) error {
	var payload NotificationUpdatePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal notification update payload: %v: %w", err, asynq.SkipRetry)
	}

	// Validate payload
	if payload.NotificationID == 0 {
		return fmt.Errorf("notification_id is required: %w", asynq.SkipRetry)
	}

	// Build update map
	updates := make(map[string]interface{})
	if payload.Type != "" {
		updates["type"] = payload.Type
	}
	if payload.Message != "" {
		updates["message"] = payload.Message
	}
	if payload.Read != nil {
		updates["read"] = *payload.Read
	}
	if payload.Title != "" {
		updates["title"] = payload.Title
	}

	if len(updates) == 0 {
		return fmt.Errorf("no valid updates provided: %w", asynq.SkipRetry)
	}

	result := database.DB.Model(&models.Notification{}).Where("id = ?", payload.NotificationID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found with id: %d: %w", payload.NotificationID, asynq.SkipRetry)
	}

	log.Printf("Notification updated successfully: id=%d", payload.NotificationID)
	return nil
}

// HandleMarkNotificationReadTask handles marking a notification as read
func HandleMarkNotificationReadTask(ctx context.Context, t *asynq.Task) error {
	var payload NotificationMarkReadPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal mark read payload: %v: %w", err, asynq.SkipRetry)
	}

	// Validate payload
	if payload.NotificationID == 0 {
		return fmt.Errorf("notification_id is required: %w", asynq.SkipRetry)
	}
	if payload.UserID == 0 {
		return fmt.Errorf("user_id is required: %w", asynq.SkipRetry)
	}

	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", payload.NotificationID, payload.UserID).
		Update("read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by user: id=%d, user_id=%d: %w",
			payload.NotificationID, payload.UserID, asynq.SkipRetry)
	}

	log.Printf("Notification marked as read: id=%d, user_id=%d", payload.NotificationID, payload.UserID)
	return nil
}

// HandleMarkAllNotificationsReadTask handles marking all notifications as read for a user
func HandleMarkAllNotificationsReadTask(ctx context.Context, t *asynq.Task) error {
	var payload NotificationMarkAllReadPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal mark all read payload: %v: %w", err, asynq.SkipRetry)
	}

	// Validate payload
	if payload.UserID == 0 {
		return fmt.Errorf("user_id is required: %w", asynq.SkipRetry)
	}

	result := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", payload.UserID, false).
		Update("read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", result.Error)
	}

	log.Printf("Marked %d notifications as read for user_id: %d", result.RowsAffected, payload.UserID)
	return nil
}

// HandleDeleteNotificationTask handles notification deletion
func HandleDeleteNotificationTask(ctx context.Context, t *asynq.Task) error {
	var payload NotificationDeletePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal delete notification payload: %v: %w", err, asynq.SkipRetry)
	}

	// Validate payload
	if payload.NotificationID == 0 {
		return fmt.Errorf("notification_id is required: %w", asynq.SkipRetry)
	}
	if payload.UserID == 0 {
		return fmt.Errorf("user_id is required: %w", asynq.SkipRetry)
	}

	result := database.DB.Where("id = ? AND user_id = ?", payload.NotificationID, payload.UserID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by user: id=%d, user_id=%d: %w",
			payload.NotificationID, payload.UserID, asynq.SkipRetry)
	}

	log.Printf("Notification deleted successfully: id=%d, user_id=%d", payload.NotificationID, payload.UserID)
	return nil
}
