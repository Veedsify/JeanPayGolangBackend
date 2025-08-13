package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// PaystackWebhookData represents Paystack webhook payload
type PaystackWebhookData struct {
	Event string `json:"event"`
	Data  struct {
		ID              int64  `json:"id"`
		Domain          string `json:"domain"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"` // Amount in kobo
		Message         string `json:"message"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		IPAddress       string `json:"ip_address"`
		Customer        struct {
			ID           int64  `json:"id"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Channel           string `json:"channel"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
			Brand             string `json:"brand"`
		} `json:"authorization"`
		Plan       interface{} `json:"plan"`
		SubAccount interface{} `json:"subaccount"`
		Log        interface{} `json:"log"`
	} `json:"data"`
}

// MomoWebhookData represents Mobile Money webhook payload
type MomoWebhookData struct {
	Event string `json:"event"`
	Data  struct {
		Reference     string  `json:"reference"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		Status        string  `json:"status"`
		TransactionID string  `json:"transaction_id"`
		PhoneNumber   string  `json:"phone_number"`
		Network       string  `json:"network"`
		ProcessedAt   string  `json:"processed_at"`
		Customer      struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Phone string `json:"phone"`
		} `json:"customer"`
	} `json:"data"`
}

type CombinedWebhookData struct {
	MomoWebhookData     *MomoWebhookData     `json:"momo,omitempty"`
	PaystackWebhookData *PaystackWebhookData `json:"paystack,omitempty"`
}

// WebhookEventLog represents webhook event for logging
type WebhookEventLog struct {
	Provider    string              `json:"provider"`
	Event       string              `json:"event"`
	Reference   string              `json:"reference"`
	Status      string              `json:"status"`
	Amount      float64             `json:"amount"`
	Currency    string              `json:"currency"`
	ProcessedAt time.Time           `json:"processed_at"`
	RawPayload  CombinedWebhookData `json:"raw_payload"`
}

// HandlePaystackWebhook processes Paystack webhook events
func HandlePaystackWebhook(payload []byte, signature string) error {
	// Verify webhook signature
	if !verifyPaystackSignature(payload, signature) {
		return errors.New("invalid webhook signature")
	}

	// Parse webhook data
	var paystackData PaystackWebhookData
	if err := json.Unmarshal(payload, &paystackData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	webhookData := CombinedWebhookData{PaystackWebhookData: &paystackData}

	// Log webhook event
	eventLog := WebhookEventLog{
		Provider:    "paystack",
		Event:       paystackData.Event,
		Reference:   paystackData.Data.Reference,
		Status:      paystackData.Data.Status,
		Amount:      float64(paystackData.Data.Amount) / 100, // Convert from kobo to naira
		Currency:    paystackData.Data.Currency,
		ProcessedAt: time.Now(),
		RawPayload:  webhookData,
	}

	// Process based on event type
	switch webhookData.PaystackWebhookData.Event {
	case "charge.success":
		return handlePaystackChargeSuccess(&webhookData, &eventLog)
	case "charge.failed":
		return handlePaystackChargeFailed(&webhookData, &eventLog)
	case "transfer.success":
		return handlePaystackTransferSuccess(&webhookData, &eventLog)
	case "transfer.failed":
		return handlePaystackTransferFailed(&webhookData, &eventLog)
	default:
		// Log unsupported event
		return logWebhookEvent(&eventLog, "unsupported_event")
	}
}

// HandleMomoWebhook processes Mobile Money webhook events
func HandleMomoWebhook(payload []byte, signature string) error {
	// Verify webhook signature (implement based on your MoMo provider)
	if !verifyMomoSignature(payload, signature) {
		return errors.New("invalid webhook signature")
	}

	var momoData MomoWebhookData
	if err := json.Unmarshal(payload, &momoData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	webhookData := CombinedWebhookData{MomoWebhookData: &momoData}

	// Log webhook event
	eventLog := WebhookEventLog{
		Provider:    "momo",
		Event:       momoData.Event,
		Reference:   momoData.Data.Reference,
		Status:      momoData.Data.Status,
		Amount:      momoData.Data.Amount,
		Currency:    momoData.Data.Currency,
		ProcessedAt: time.Now(),
		RawPayload:  webhookData,
	}

	// Process based on event type
	switch webhookData.MomoWebhookData.Event {
	case "payment.success":
		return handleMomoPaymentSuccess(&webhookData, &eventLog)
	case "payment.failed":
		return handleMomoPaymentFailed(&webhookData, &eventLog)
	case "transfer.success":
		return handleMomoTransferSuccess(&webhookData, &eventLog)
	case "transfer.failed":
		return handleMomoTransferFailed(&webhookData, &eventLog)
	default:
		// Log unsupported event
		return logWebhookEvent(&eventLog, "unsupported_event")
	}
}

// handlePaystackChargeSuccess processes successful Paystack charges (deposits)
func handlePaystackChargeSuccess(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.PaystackWebhookData.Data.Reference
	amount := float64(webhookData.PaystackWebhookData.Data.Amount) / 100 // Convert kobo to naira

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Transaction not found, might be external payment
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Check if already processed
	if transaction.Status == "completed" {
		return logWebhookEvent(eventLog, "already_processed")
	}

	// Start database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	if err := tx.Model(&transaction).Updates(map[string]interface{}{
		"status":     "completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update wallet balance
	if transaction.TransactionType == "deposit" {
		err := updateWalletBalance(tx, transaction.UserID, transaction.TransactionDetails.FromCurrency, amount, "deposit")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update wallet: %w", err)
		}
	}

	// Create notification
	if err := createTransactionNotification(tx, transaction.UserID, "deposit", amount, transaction.TransactionDetails.FromCurrency, transaction.TransactionID); err != nil {
		// Log error but don't fail the transaction
		fmt.Printf("Failed to create notification: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handlePaystackChargeFailed processes failed Paystack charges
func handlePaystackChargeFailed(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.PaystackWebhookData.Data.Reference

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Update transaction status
	if err := database.DB.Model(&transaction).Updates(map[string]interface{}{
		"status":      "failed",
		"description": transaction.Description + " | Payment failed: " + webhookData.PaystackWebhookData.Data.Message,
		"updated_at":  time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handlePaystackTransferSuccess processes successful Paystack transfers (withdrawals)
func handlePaystackTransferSuccess(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.PaystackWebhookData.Data.Reference

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Update transaction status
	if err := database.DB.Model(&transaction).Updates(map[string]interface{}{
		"status":     "completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Create notification
	amount := float64(webhookData.PaystackWebhookData.Data.Amount) / 100
	createTransactionNotificationDirect(transaction.UserID, "withdrawal", amount, transaction.TransactionDetails.FromCurrency, transaction.TransactionID)

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handlePaystackTransferFailed processes failed Paystack transfers
func handlePaystackTransferFailed(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.PaystackWebhookData.Data.Reference
	amount := float64(webhookData.PaystackWebhookData.Data.Amount) / 100

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Start database transaction for refund
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	if err := tx.Model(&transaction).Updates(map[string]interface{}{
		"status":      "failed",
		"description": transaction.Description + " | Transfer failed: " + webhookData.PaystackWebhookData.Data.Message,
		"updated_at":  time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Refund wallet balance for failed withdrawal
	if transaction.TransactionType == "withdrawal" {
		err := updateWalletBalance(tx, transaction.UserID, transaction.TransactionDetails.FromCurrency, amount, "refund")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund wallet: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handleMomoPaymentSuccess processes successful MoMo payments
func handleMomoPaymentSuccess(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.MomoWebhookData.Data.Reference
	amount := webhookData.MomoWebhookData.Data.Amount

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Check if already processed
	if transaction.Status == "completed" {
		return logWebhookEvent(eventLog, "already_processed")
	}

	// Start database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	if err := tx.Model(&transaction).Updates(map[string]interface{}{
		"status":     "completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update wallet balance for deposits
	if transaction.TransactionType == "deposit" {
		err := updateWalletBalance(tx, transaction.UserID, transaction.TransactionDetails.FromCurrency, amount, "deposit")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update wallet: %w", err)
		}
	}

	// Create notification
	if err := createTransactionNotification(tx, transaction.UserID, "deposit", amount, transaction.TransactionDetails.FromCurrency, transaction.TransactionID); err != nil {
		fmt.Printf("Failed to create notification: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handleMomoPaymentFailed processes failed MoMo payments
func handleMomoPaymentFailed(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.MomoWebhookData.Data.Reference

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Update transaction status
	if err := database.DB.Model(&transaction).Updates(map[string]interface{}{
		"status":      "failed",
		"description": transaction.Description + " | MoMo payment failed",
		"updated_at":  time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handleMomoTransferSuccess processes successful MoMo transfers
func handleMomoTransferSuccess(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.MomoWebhookData.Data.Reference

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Update transaction status
	if err := database.DB.Model(&transaction).Updates(map[string]interface{}{
		"status":     "completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Create notification
	createTransactionNotificationDirect(transaction.UserID, "withdrawal", webhookData.MomoWebhookData.Data.Amount, transaction.TransactionDetails.FromCurrency, transaction.TransactionID)

	return logWebhookEvent(eventLog, "processed_successfully")
}

// handleMomoTransferFailed processes failed MoMo transfers
func handleMomoTransferFailed(webhookData *CombinedWebhookData, eventLog *WebhookEventLog) error {
	reference := webhookData.MomoWebhookData.Data.Reference
	amount := webhookData.MomoWebhookData.Data.Amount

	// Find transaction by reference
	var transaction models.Transaction
	err := database.DB.Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return logWebhookEvent(eventLog, "transaction_not_found")
		}
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	// Start database transaction for refund
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update transaction status
	if err := tx.Model(&transaction).Updates(map[string]interface{}{
		"status":      "failed",
		"description": transaction.Description + " | MoMo transfer failed",
		"updated_at":  time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Refund wallet balance for failed withdrawal
	if transaction.TransactionType == "withdrawal" {
		err := updateWalletBalance(tx, transaction.UserID, transaction.TransactionDetails.FromCurrency, amount, "refund")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund wallet: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return logWebhookEvent(eventLog, "processed_successfully")
}

// Helper functions

// verifyPaystackSignature verifies Paystack webhook signature
func verifyPaystackSignature(payload []byte, signature string) bool {
	secret := []byte(libs.GetEnvOrDefault("PAYSTACK_SECRET_KEY", ""))
	if len(secret) == 0 {
		return false
	}

	h := hmac.New(sha512.New, secret)
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// verifyMomoSignature verifies Mobile Money webhook signature
func verifyMomoSignature(payload []byte, signature string) bool {
	// Implement based on your MoMo provider's signature verification
	secret := []byte(libs.GetEnvOrDefault("MOMO_SECRET_KEY", ""))
	if len(secret) == 0 {
		return false
	}

	h := hmac.New(sha512.New, secret)
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// logWebhookEvent logs webhook processing events
func logWebhookEvent(eventLog *WebhookEventLog, result string) error {
	webhookEvent := models.WebhookEvent{
		EventID:   eventLog.Reference + "_" + strconv.FormatInt(time.Now().Unix(), 10),
		EventType: eventLog.Event,
		Provider:  eventLog.Provider,
		Payload:   eventLog.RawPayload,
		Status:    models.WebhookEventStatus(result),
	}

	if err := database.DB.Create(&webhookEvent).Error; err != nil {
		return fmt.Errorf("failed to log webhook event: %w", err)
	}

	return nil
}

// createTransactionNotification creates notification within a transaction
func createTransactionNotification(tx *gorm.DB, userID uint, txType string, amount float64, currency, transactionID string) error {
	message := fmt.Sprintf("Your %s of %s %.2f has been processed successfully. Transaction ID: %s",
		txType, currency, amount, transactionID)

	notification := models.Notification{
		UserID:  userID,
		Type:    "transaction_" + txType,
		Message: message,
		Read:    false,
	}

	return tx.Create(&notification).Error
}

// createTransactionNotificationDirect creates notification directly
func createTransactionNotificationDirect(userID uint, txType string, amount float64, currency, transactionID string) error {
	message := fmt.Sprintf("Your %s of %s %.2f has been processed successfully. Transaction ID: %s",
		txType, currency, amount, transactionID)

	notification := models.Notification{
		UserID:  userID,
		Type:    "transaction_" + txType,
		Message: message,
		Read:    false,
	}

	return database.DB.Create(&notification).Error
}

// GetWebhookEventLogs retrieves webhook event logs for admin
func GetWebhookEventLogs(pagination types.PaginationRequest, provider, status string) ([]models.WebhookEvent, *types.PaginationResponse, error) {
	query := database.DB.Model(&models.WebhookEvent{})

	// Apply filters
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to count webhook events: %w", err)
	}

	// Calculate pagination
	page, limit := pagination.GetValidatedParams()
	offset := (page - 1) * limit

	// Find events
	var events []models.WebhookEvent
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find webhook events: %w", err)
	}

	// Create pagination response
	paginationResp := types.NewPaginationResponse(page, limit, total)

	return events, paginationResp, nil
}

// Get total balance for a user

func GetUserTotalBalance(userID uint) (float64, error) {
	var totalBalance float64
	var user models.User

	if err := database.DB.Preload("Wallet").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("user not found with ID: %d", userID)
		}
		return 0, fmt.Errorf("failed to find user: %w", err)
	}
	userCountry := user.Country
	switch userCountry {
		case "NG":
			for _, wallet := range user.Wallet {
				if wallet.Currency == "NGN" {
					totalBalance += wallet.Balance
				}
			}
		case "GH":
			for _, wallet := range user.Wallet {
				if wallet.Currency == "GHS" {
					totalBalance += wallet.Balance
				}
			}
	}

	return totalBalance, nil
}
