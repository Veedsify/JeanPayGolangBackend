package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/interfaces"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/hibiken/asynq"
)

var redisAddr = libs.GetEnvOrDefault("REDIS_ADDR", "127.0.0.1:6379")

// Email job types
const (
	TypeEmailDelivery       = "email:delivery"
	TypeWelcomeEmail        = "email:welcome"
	TypePasswordResetEmail  = "email:password_reset"
	TypeEmailVerification   = "email:verification"
	TypeTwoFactorEmail      = "email:two_factor_authentication"
	TypeTransactionApproved = "email:transaction_approved"
	TypeTransactionRejected = "email:transaction_rejected"
)

// Base email job payload
type EmailJobPayload struct {
	To          []string       `json:"to"`
	Subject     string         `json:"subject"`
	TemplateID  string         `json:"template_id"`
	Data        map[string]any `json:"data"`
	Priority    string         `json:"priority,omitempty"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty"`
}

// Specific email job payloads
type WelcomeEmailPayload struct {
	EmailJobPayload
	UserName string `json:"user_name"`
	Token    string `json:"token"`
}

type PasswordResetEmailPayload struct {
	EmailJobPayload
	ResetToken string `json:"reset_token"`
}

type EmailVerificationPayload struct {
	EmailJobPayload
	UserName          string `json:"user_name"`
	VerificationToken string `json:"verification_token"`
}

type TransactionApprovedPayload struct {
	EmailJobPayload
	UserName        string             `json:"user_name"`
	TransactionType string             `json:"transaction_type"`
	Amount          string             `json:"amount"`
	TransactionID   string             `json:"transaction_id"`
	Transaction     models.Transaction `json:"transaction"`
}

type TransactionRejectedPayload struct {
	EmailJobPayload
	UserName        string             `json:"user_name"`
	TransactionType string             `json:"transaction_type"`
	Amount          string             `json:"amount"`
	TransactionID   string             `json:"transaction_id"`
	Reason          string             `json:"reason"`
	Transaction     models.Transaction `json:"transaction"`
}

// EmailJobClient handles email job creation and queuing
type EmailJobClient struct {
	client *asynq.Client
}

// NewEmailJobClient creates a new email job client
func NewEmailJobClient() *EmailJobClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &EmailJobClient{
		client: client,
	}
}

// Close closes the email job client
func (ejc *EmailJobClient) Close() error {
	return ejc.client.Close()
}

// EnqueueWelcomeEmail queues a welcome email job
func (ejc *EmailJobClient) EnqueueWelcomeEmail(email, userName, token string) error {
	payload := WelcomeEmailPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    "Welcome to JeanPay!",
			TemplateID: "welcome",
			Data: map[string]any{
				"user_name": userName,
				"token":     token,
			},
			Priority: "high",
		},
		UserName: userName,
		Token:    token,
	}

	task, err := createEmailTask(TypeWelcomeEmail, payload)
	if err != nil {
		return fmt.Errorf("failed to create welcome email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue welcome email task: %w", err)
	}

	log.Printf("Enqueued welcome email task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueuePasswordResetEmail queues a password reset email job
func (ejc *EmailJobClient) EnqueuePasswordResetEmail(email, resetToken string) error {
	payload := PasswordResetEmailPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    "Password Reset Request - JeanPay",
			TemplateID: "password_reset",
			Data: map[string]any{
				"reset_token": resetToken,
			},
			Priority: "high",
		},
		ResetToken: resetToken,
	}

	task, err := createEmailTask(TypePasswordResetEmail, payload)
	if err != nil {
		return fmt.Errorf("failed to create password reset email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue password reset email task: %w", err)
	}

	log.Printf("Enqueued password reset email task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueEmailVerification queues an email verification job
func (ejc *EmailJobClient) EnqueueEmailVerification(email, userName, verificationToken string) error {
	payload := EmailVerificationPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    "Verify Your Email - JeanPay",
			TemplateID: "email_verification",
			Data: map[string]any{
				"user_name":          userName,
				"verification_token": verificationToken,
			},
			Priority: "high",
		},
		UserName:          userName,
		VerificationToken: verificationToken,
	}

	task, err := createEmailTask(TypeEmailVerification, payload)
	if err != nil {
		return fmt.Errorf("failed to create email verification task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue email verification task: %w", err)
	}

	log.Printf("Enqueued email verification task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueTwoFactorEmail queues a two-factor authentication email job
func (ejc *EmailJobClient) EnqueueTwoFactorEmail(email, userName, verificationLink string) error {
	payload := EmailJobPayload{
		To:         []string{email},
		Subject:    "Two-Factor Authentication Code - JeanPay",
		TemplateID: "two_factor_auth",
		Data: map[string]any{
			"user_name":         userName,
			"verification_link": verificationLink,
		},
		Priority: "high",
	}
	task, err := createEmailTask(TypeTwoFactorEmail, payload)
	if err != nil {
		return fmt.Errorf("failed to create two-factor email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue two-factor email task: %w", err)
	}
	log.Printf("Enqueued two-factor email task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueTransactionApproved queues an approved transaction email job
func (ejc *EmailJobClient) EnqueueTransactionApproved(email string, userName string, transaction models.Transaction) error {
	payload := TransactionApprovedPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    fmt.Sprintf("%s Approved - JeanPay", getTransactionTypeDisplay(string(transaction.TransactionType))),
			TemplateID: "transaction_approved",
			Data: map[string]any{
				"user_name":        userName,
				"transaction_id":   transaction.TransactionID,
				"transaction_type": transaction.TransactionType,
				"amount":           utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
				"email":            email,
			},
			Priority: "high",
		},
		UserName:        userName,
		TransactionType: string(transaction.TransactionType),
		Amount:          utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		TransactionID:   transaction.TransactionID,
		Transaction:     transaction,
	}

	task, err := createEmailTask(TypeTransactionApproved, payload)
	if err != nil {
		return fmt.Errorf("failed to create transaction approved email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue transaction approved email task: %w", err)
	}

	log.Printf("Enqueued transaction approved email task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// EnqueueTransactionRejected queues a rejected transaction email job
func (ejc *EmailJobClient) EnqueueTransactionRejected(email string, userName string, transaction models.Transaction, reason string) error {
	payload := TransactionRejectedPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    fmt.Sprintf("%s Rejected - JeanPay", getTransactionTypeDisplay(string(transaction.TransactionType))),
			TemplateID: "transaction_rejected",
			Data: map[string]any{
				"user_name":        userName,
				"transaction_id":   transaction.TransactionID,
				"transaction_type": transaction.TransactionType,
				"amount":           utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
				"reason":           reason,
				"email":            email,
			},
			Priority: "high",
		},
		UserName:        userName,
		TransactionType: string(transaction.TransactionType),
		Amount:          utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		TransactionID:   transaction.TransactionID,
		Reason:          reason,
		Transaction:     transaction,
	}

	task, err := createEmailTask(TypeTransactionRejected, payload)
	if err != nil {
		return fmt.Errorf("failed to create transaction rejected email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("high"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue transaction rejected email task: %w", err)
	}

	log.Printf("Enqueued transaction rejected email task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

// Helper function to get user-friendly transaction type display names
func getTransactionTypeDisplay(transactionType string) string {
	switch transactionType {
	case "deposit":
		return "Deposit"
	case "withdrawal":
		return "Withdrawal"
	case "transfer":
		return "Transfer"
	case "conversion":
		return "Currency Conversion"
	default:
		return "Transaction"
	}
}

// EnqueueScheduledEmail queues an email to be sent at a specific time
func (ejc *EmailJobClient) EnqueueScheduledEmail(payload EmailJobPayload, scheduledAt time.Time) error {
	payload.ScheduledAt = &scheduledAt

	task, err := createEmailTask(TypeEmailDelivery, payload)
	if err != nil {
		return fmt.Errorf("failed to create scheduled email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.ProcessAt(scheduledAt),
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue scheduled email task: %w", err)
	}

	log.Printf("Enqueued scheduled email task: id=%s queue=%s scheduled_at=%s", info.ID, info.Queue, scheduledAt.Format(time.RFC3339))
	return nil
}

// createEmailTask creates an asynq task for email delivery
func createEmailTask(taskType string, payload any) (*asynq.Task, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(taskType, payloadBytes), nil
}

// Worker functions

// HandleWelcomeEmailTask handles welcome email delivery
func HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload WelcomeEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal welcome email payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid welcome email payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendWelcomeEmail(payload.To[0], payload.UserName, payload.Token)
	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	log.Printf("Welcome email sent successfully to: %s", payload.To[0])
	return nil
}

// HandleTwoFactorEmailTask handles two-factor authentication email delivery

func HandleTwoFactorEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal two-factor email payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload); err != nil {
		return fmt.Errorf("invalid two-factor email payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendTwoFactorAuthenticationEmail(payload.To[0], payload.Data["user_name"].(string), payload.Data["verification_link"].(string))
	if err != nil {
		return fmt.Errorf("failed to send two-factor email: %w", err)
	}

	log.Printf("Two-factor email sent successfully to: %s", payload.To[0])
	return nil
}

// HandlePasswordResetEmailTask handles password reset email delivery
func HandlePasswordResetEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload PasswordResetEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal password reset email payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid password reset email payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendPasswordResetEmail(payload.To[0], payload.ResetToken)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	log.Printf("Password reset email sent successfully to: %s", payload.To[0])
	return nil
}

// HandleEmailVerificationTask handles email verification delivery
func HandleEmailVerificationTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailVerificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email verification payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid email verification payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendEmailVerification(
		payload.To[0],
		payload.UserName,
		payload.VerificationToken,
	)
	if err != nil {
		return fmt.Errorf("failed to send email verification: %w", err)
	}

	log.Printf("Email verification sent successfully to: %s", payload.To[0])
	return nil
}

// HandleTransactionApprovedTask handles transaction approved email delivery
func HandleTransactionApprovedTask(ctx context.Context, t *asynq.Task) error {
	var payload TransactionApprovedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal transaction approved payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid transaction approved payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendTransactionApprovedEmail(
		payload.To[0],
		payload.UserName,
		payload.Transaction,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction approved email: %w", err)
	}

	log.Printf("Transaction approved email sent successfully to: %s", payload.To[0])
	return nil
}

// HandleTransactionRejectedTask handles transaction rejected email delivery
func HandleTransactionRejectedTask(ctx context.Context, t *asynq.Task) error {
	var payload TransactionRejectedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal transaction rejected payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid transaction rejected payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendTransactionRejectedEmail(
		payload.To[0],
		payload.UserName,
		payload.Transaction,
		payload.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction rejected email: %w", err)
	}

	log.Printf("Transaction rejected email sent successfully to: %s", payload.To[0])
	return nil
}

// HandleGenericEmailTask handles generic email delivery
func HandleGenericEmailTask(ctx context.Context, t *asynq.Task) error {
	var payload EmailJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload); err != nil {
		return fmt.Errorf("invalid email payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	var err error
	if payload.TemplateID != "" && payload.Data != nil {
		err = emailSender.SendTemplatedEmail(payload.To, payload.TemplateID, payload.Data)
	} else {
		body := fmt.Sprintf("Template: %s\nData: %+v", payload.TemplateID, payload.Data)
		err = emailSender.SendSimpleEmail(payload.To, payload.Subject, body)
	}

	if err != nil {
		return fmt.Errorf("failed to send generic email: %w", err)
	}

	log.Printf("Generic email sent successfully to: %v", payload.To)
	return nil
}

// validateEmailPayload validates the email job payload
func validateEmailPayload(payload EmailJobPayload) error {
	if len(payload.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if payload.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if payload.TemplateID == "" {
		return fmt.Errorf("template_id is required")
	}

	// Validate email addresses
	if slices.Contains(payload.To, "") {
		return fmt.Errorf("empty email address in recipients")
	}

	return nil
}
