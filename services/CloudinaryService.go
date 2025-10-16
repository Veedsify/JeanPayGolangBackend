package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService() (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, err
	}
	return &CloudinaryService{cld: cld}, nil
}

func (s *CloudinaryService) UploadImage(file multipart.File, folder string) (string, error) {
	uploadParams := uploader.UploadParams{
		Folder: folder,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", err
	}
	return result.SecureURL, nil
}

// UploadReceipt uploads a receipt file to Cloudinary with transaction ID folder organization
func (s *CloudinaryService) UploadReceipt(file multipart.File, header *multipart.FileHeader, transactionID string) (string, error) {
	// Validate file type
	if err := s.ValidateReceiptFile(header); err != nil {
		return "", err
	}

	// Create folder structure: receipts/transactionID
	folder := fmt.Sprintf("receipts/%s", transactionID)
	
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "auto", // Allow images and PDFs
		PublicID:     fmt.Sprintf("receipt_%d", time.Now().Unix()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload receipt to Cloudinary: %w", err)
	}
	
	return result.SecureURL, nil
}

// ValidateReceiptFile validates file type and size for receipt uploads
func (s *CloudinaryService) ValidateReceiptFile(header *multipart.FileHeader) error {
	// Check file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxFileSize {
		return fmt.Errorf("file size exceeds maximum limit of 10MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("invalid file type. Only .jpg, .jpeg, .png, and .pdf files are allowed")
	}

	// Check MIME type
	contentType := header.Header.Get("Content-Type")
	allowedMimeTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedMimeTypes[contentType] {
		return fmt.Errorf("invalid file type. Content-Type must be image/jpeg, image/png, or application/pdf")
	}

	return nil
}
