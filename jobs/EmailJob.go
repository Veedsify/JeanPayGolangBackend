package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/interfaces"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/hibiken/asynq"
)

var redisAddr = libs.GetEnvOrDefault("REDIS_ADDR", "127.0.0.1:6379")

// Email job types
const (
	TypeEmailDelivery           = "email:delivery"
	TypeWelcomeEmail            = "email:welcome"
	TypePasswordResetEmail      = "email:password_reset"
	TypeTransactionNotification = "email:transaction_notification"
	TypeEmailVerification       = "email:verification"
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

type TransactionNotificationPayload struct {
	EmailJobPayload
	TransactionType string `json:"transaction_type"`
	Amount          string `json:"amount"`
	TransactionID   string `json:"transaction_id"`
}

type EmailVerificationPayload struct {
	EmailJobPayload
	UserName          string `json:"user_name"`
	VerificationToken string `json:"verification_token"`
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

// EnqueueTransactionNotification queues a transaction notification email job
func (ejc *EmailJobClient) EnqueueTransactionNotification(email, transactionType, amount, transactionID string) error {
	payload := TransactionNotificationPayload{
		EmailJobPayload: EmailJobPayload{
			To:         []string{email},
			Subject:    fmt.Sprintf("Transaction %s - JeanPay", transactionType),
			TemplateID: "transaction_notification",
			Data: map[string]any{
				"transaction_type": transactionType,
				"amount":           amount,
				"transaction_id":   transactionID,
			},
			Priority: "medium",
		},
		TransactionType: transactionType,
		Amount:          amount,
		TransactionID:   transactionID,
	}

	task, err := createEmailTask(TypeTransactionNotification, payload)
	if err != nil {
		return fmt.Errorf("failed to create transaction notification email task: %w", err)
	}

	opts := []asynq.Option{
		asynq.Queue("default"),
		asynq.MaxRetry(2),
		asynq.Timeout(3 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue transaction notification email task: %w", err)
	}

	log.Printf("Enqueued transaction notification email task: id=%s queue=%s", info.ID, info.Queue)
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

// HandleTransactionNotificationTask handles transaction notification email delivery
func HandleTransactionNotificationTask(ctx context.Context, t *asynq.Task) error {
	var payload TransactionNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal transaction notification payload: %v: %w", err, asynq.SkipRetry)
	}

	if err := validateEmailPayload(payload.EmailJobPayload); err != nil {
		return fmt.Errorf("invalid transaction notification payload: %v: %w", err, asynq.SkipRetry)
	}

	emailSender := interfaces.GetGlobalEmailSender()
	if emailSender == nil {
		return fmt.Errorf("email sender not initialized")
	}

	err := emailSender.SendTransactionNotification(
		payload.To[0],
		payload.TransactionType,
		payload.Amount,
		payload.TransactionID,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction notification email: %w", err)
	}

	log.Printf("Transaction notification email sent successfully to: %s", payload.To[0])
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
