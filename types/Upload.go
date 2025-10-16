package types

import "time"

// UploadReceiptRequest represents the request for uploading a receipt
type UploadReceiptRequest struct {
	TransactionID string `form:"transaction_id" binding:"required"`
	Receipt       string `form:"receipt" binding:"required"` // This will be handled as multipart file
}

// UploadReceiptResponse represents the response after uploading a receipt
type UploadReceiptResponse struct {
	ReceiptID     uint      `json:"receipt_id"`
	CloudinaryURL string    `json:"cloudinary_url"`
	TransactionID string    `json:"transaction_id"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

// GetReceiptResponse represents the response when retrieving receipt information
type GetReceiptResponse struct {
	ID            uint      `json:"id"`
	TransactionID string    `json:"transaction_id"`
	CloudinaryURL string    `json:"cloudinary_url"`
	OriginalName  string    `json:"original_name"`
	FileSize      int64     `json:"file_size"`
	FileType      string    `json:"file_type"`
	UploadedAt    time.Time `json:"uploaded_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}