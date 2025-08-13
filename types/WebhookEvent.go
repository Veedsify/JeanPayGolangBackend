package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateWebhookEventRequest struct {
	EventType string                 `json:"event_type" form:"event_type" binding:"required"`
	Provider  string                 `json:"provider" form:"provider" binding:"required"`
	Payload   map[string]interface{} `json:"payload" form:"payload" binding:"required"`
}

type UpdateWebhookEventRequest struct {
	Status models.WebhookEventStatus `json:"status" form:"status"`
}

type WebhookEventResponse struct {
	ID        uint32                    `json:"id"`
	EventID   string                    `json:"event_id"`
	EventType string                    `json:"event_type"`
	Provider  string                    `json:"provider"`
	Payload   interface{}               `json:"payload"`
	Status    models.WebhookEventStatus `json:"status"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

type GetWebhookEventsRequest struct {
	EventType string `form:"event_type"`
	Provider  string `form:"provider"`
	Status    string `form:"status"`
	FromDate  string `form:"from_date"`
	ToDate    string `form:"to_date"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

type GetWebhookEventsResponse struct {
	Events     []WebhookEventResponse `json:"events"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

type WebhookEventStatsResponse struct {
	TotalEvents           int64                  `json:"total_events"`
	ByStatus              map[string]int64       `json:"by_status"`
	ByProvider            map[string]int64       `json:"by_provider"`
	ByEventType           map[string]int64       `json:"by_event_type"`
	RecentEvents          []WebhookEventResponse `json:"recent_events"`
	ProcessingRate        float64                `json:"processing_rate"`
	FailureRate           float64                `json:"failure_rate"`
	AverageProcessingTime string                 `json:"average_processing_time"`
}

type RetryWebhookEventRequest struct {
	EventID string `json:"event_id" form:"event_id" binding:"required"`
}

type RetryWebhookEventResponse struct {
	EventID    string                    `json:"event_id"`
	Status     models.WebhookEventStatus `json:"status"`
	RetriedAt  time.Time                 `json:"retried_at"`
	RetryCount int                       `json:"retry_count"`
	Message    string                    `json:"message"`
}

type ProcessWebhookRequest struct {
	EventID   string                 `json:"event_id" form:"event_id" binding:"required"`
	EventType string                 `json:"event_type" form:"event_type" binding:"required"`
	Provider  string                 `json:"provider" form:"provider" binding:"required"`
	Payload   map[string]interface{} `json:"payload" form:"payload" binding:"required"`
	Signature string                 `json:"signature" form:"signature"`
}

type ProcessWebhookResponse struct {
	EventID     string                    `json:"event_id"`
	Status      models.WebhookEventStatus `json:"status"`
	ProcessedAt time.Time                 `json:"processed_at"`
	Message     string                    `json:"message"`
	Success     bool                      `json:"success"`
}

func ToWebhookEventResponse(event *models.WebhookEvent) WebhookEventResponse {
	return WebhookEventResponse{
		ID:        uint32(event.ID),
		EventID:   event.EventID,
		EventType: event.EventType,
		Provider:  event.Provider,
		Payload:   event.Payload,
		Status:    event.Status,
		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}
}

func ToWebhookEventsResponse(events []models.WebhookEvent) []WebhookEventResponse {
	var response []WebhookEventResponse
	for _, event := range events {
		response = append(response, ToWebhookEventResponse(&event))
	}
	return response
}
