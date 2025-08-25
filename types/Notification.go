package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateNotificationRequest struct {
	UserID  uint                    `json:"user_id" form:"user_id" binding:"required"`
	Type    models.NotificationType `json:"type" form:"type" binding:"required"`
	Title   string                  `json:"title" form:"title" binding:"required"`
	Message string                  `json:"message" form:"message" binding:"required"`
	Data    string                  `json:"data" form:"data"`
}

type UpdateNotificationRequest struct {
	Read *bool `json:"read" form:"read"`
}

type NotificationResponse struct {
	ID        uint                    `json:"id"`
	UserID    uint                    `json:"user_id"`
	Type      models.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Read      bool                    `json:"read"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type GetNotificationsRequest struct {
	UserID string `form:"user_id"`
	Type   string `form:"type"`
	Read   *bool  `form:"read"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type GetNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	NextCursor    *uint                  `json:"nextCursor"`
	UnreadCount   int64                  `json:"unread_count"`
}

type MarkAllAsReadRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required"`
}

type NotificationStatsResponse struct {
	Total     int64 `json:"total"`
	Read      int64 `json:"read"`
	Unread    int64 `json:"unread"`
	LastWeek  int64 `json:"last_week"`
	LastMonth int64 `json:"last_month"`
}

func ToNotificationResponse(notification *models.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Read:      notification.Read,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

type NotificationQuery struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	ReadStatus string `form:"read_status"`
	Cursor     *uint  `form:"cursor"`
	Type       string `form:"type"`
}

type NotificationDetailsResponse struct {
	ID        uint32                  `json:"id"`
	UserID    string                  `json:"user_id"`
	Type      models.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Data      string                  `json:"data"`
	Read      bool                    `json:"read"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type CreateNotificationResponse struct {
	ID        uint32                  `json:"id"`
	UserID    string                  `json:"user_id"`
	Type      models.NotificationType `json:"type"`
	Title     string                  `json:"title"`
	Message   string                  `json:"message"`
	Read      bool                    `json:"read"`
	CreatedAt time.Time               `json:"created_at"`
}

func ToNotificationsResponse(notifications []models.Notification) []NotificationResponse {
	var response []NotificationResponse
	for _, notification := range notifications {
		response = append(response, ToNotificationResponse(&notification))
	}
	return response
}
