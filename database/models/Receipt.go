package models

import (
	"time"

	"gorm.io/gorm"
)

type Receipt struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	TransactionID uint           `json:"transaction_id" gorm:"not null;index"`
	CloudinaryURL string         `json:"cloudinary_url" gorm:"not null"`
	OriginalName  string         `json:"original_name"`
	FileSize      int64          `json:"file_size"`
	FileType      string         `json:"file_type"`
	UploadedAt    time.Time      `json:"uploaded_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationship will be handled from Transaction side
}

func (Receipt) TableName() string {
	return "receipts"
}
