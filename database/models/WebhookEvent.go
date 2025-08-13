package models

import (
	"gorm.io/gorm"
)

type WebhookEventStatus string

const (
	WebhookPending   WebhookEventStatus = "pending"
	WebhookProcessed WebhookEventStatus = "processed"
	WebhookFailed    WebhookEventStatus = "failed"
)

type WebhookEvent struct {
	gorm.Model
	EventID   string             `json:"event_id" gorm:"not null;uniqueIndex"`
	EventType string             `json:"event_type" gorm:"not null"`
	Provider  string             `json:"provider" gorm:"not null"`
	Payload   interface{}        `json:"payload" gorm:"type:jsonb;not null"`
	Status    WebhookEventStatus `json:"status" gorm:"default:pending"`
}

func (WebhookEvent) TableName() string {
	return "webhook_events"
}
