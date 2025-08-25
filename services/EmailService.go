package services

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"

	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/interfaces"
	"github.com/Veedsify/JeanPayGoBackend/templates"
	"github.com/Veedsify/JeanPayGoBackend/utils"
)

// EmailConfig holds the SMTP server configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	UseTLS       bool
	InsecureTLS  bool
	Timeout      time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	PoolSize     int
	EnableLogger bool
}

// EmailMessage represents an email to be sent
type EmailMessage struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []EmailAttachment
	Headers     map[string]string
	Priority    EmailPriority
}

// EmailAttachment represents a file attachment
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
	Inline      bool
	ContentID   string
}

// EmailPriority represents email priority levels
type EmailPriority int

const (
	PriorityLow EmailPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	Name        string
	Subject     string
	HTMLContent string
	TextContent string
}

// EmailService handles email operations
type EmailService struct {
	config    *EmailConfig
	templates map[string]*EmailTemplate
	logger    *log.Logger
}

// EmailValidationResult represents the result of email validation
type EmailValidationResult struct {
	IsValid bool
	Error   string
}

var (
	// SERVER is the base URL for the application
	SERVER   = GetEnvOrDefault("SERVER_URL", "http://localhost:8080")
	FRONTEND = GetEnvOrDefault("FRONTEND_URL", "http://localhost:3000")

	// emailRegex is used for basic email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// GetEnvOrDefault gets environment variable or returns default value
func GetEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// NewEmailService creates a new email service instance
func NewEmailService(config *EmailConfig) *EmailService {
	if config == nil {
		config = &EmailConfig{}
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}

	service := &EmailService{
		config:    config,
		templates: make(map[string]*EmailTemplate),
	}

	// Setup logger
	if config.EnableLogger {
		service.logger = log.New(os.Stdout, "[EmailService] ", log.LstdFlags|log.Lshortfile)
	}

	// Load default templates
	service.loadDefaultTemplates()

	return service
}

// NewEmailServiceFromEnv creates email service from environment variables
func NewEmailServiceFromEnv() (*EmailService, error) {
	config := &EmailConfig{
		SMTPHost:     GetEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvIntOrDefault("SMTP_PORT", 587),
		Username:     os.Getenv("SMTP_USERNAME"),
		Password:     os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     GetEnvOrDefault("FROM_NAME", "JeanPay"),
		UseTLS:       getEnvBoolOrDefault("SMTP_USE_TLS", true),
		InsecureTLS:  getEnvBoolOrDefault("SMTP_INSECURE_TLS", false),
		Timeout:      time.Duration(getEnvIntOrDefault("SMTP_TIMEOUT", 30)) * time.Second,
		MaxRetries:   getEnvIntOrDefault("SMTP_MAX_RETRIES", 3),
		RetryDelay:   time.Duration(getEnvIntOrDefault("SMTP_RETRY_DELAY", 1)) * time.Second,
		PoolSize:     getEnvIntOrDefault("SMTP_POOL_SIZE", 10),
		EnableLogger: getEnvBoolOrDefault("SMTP_ENABLE_LOGGER", true),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email configuration: %w", err)
	}

	return NewEmailService(config), nil
}

// Validate validates the email configuration
func (c *EmailConfig) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("SMTP_USERNAME is required")
	}
	if c.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required")
	}
	if c.FromEmail == "" {
		return fmt.Errorf("FROM_EMAIL is required")
	}
	if !IsValidEmail(c.FromEmail) {
		return fmt.Errorf("FROM_EMAIL is not a valid email address")
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTP_PORT must be between 1 and 65535")
	}
	return nil
}

// loadDefaultTemplates loads the default email templates
func (es *EmailService) loadDefaultTemplates() {

	// --- Welcome Template ---
	es.templates["welcome"] = &EmailTemplate{
		Name:        "welcome",
		Subject:     "🎉 Welcome to JeanPay - Your Premium Payment Experience Awaits!",
		HTMLContent: templates.WelcomeTemplate(),
		TextContent: templates.WelcomePlainTextTemplate(),
	}

	// --- Password Reset Template ---
	es.templates["password_reset"] = &EmailTemplate{
		Name:        "password_reset",
		Subject:     "🔐 Secure Password Reset Request - JeanPay",
		HTMLContent: templates.PasswordResetTemplate(),
		TextContent: templates.PasswordResetPlainTextTemplate(),
	}

	// --- Two-Factor Authentication Template ---
	es.templates["two_factor_authentication"] = &EmailTemplate{
		Name:        "two_factor_authentication",
		Subject:     "🔐 Your JeanPay Security Code - Expires in 10 Minutes",
		HTMLContent: templates.TwoFactorAuthenticationTemplate(),
		TextContent: templates.TwoFactorAuthenticationPlainTemplate(),
	}

	// --- Transaction Approved Template ---
	es.templates["transaction_approved"] = &EmailTemplate{
		Name:        "transaction_approved",
		Subject:     "✅ {{.TransactionTypeDisplay}} Approved - JeanPay",
		HTMLContent: templates.TransactionApprovedTemplate(),
		TextContent: templates.TransactionApprovedPlainTextTemplate(),
	}

	// --- Transaction Rejected Template ---
	es.templates["transaction_rejected"] = &EmailTemplate{
		Name:        "transaction_rejected",
		Subject:     "❌ {{.TransactionTypeDisplay}} Rejected - JeanPay",
		HTMLContent: templates.TransactionRejectedTemplate(),
		TextContent: templates.TransactionRejectedPlainTextTemplate(),
	}
}

// SendEmail sends an email message with retry logic
func (es *EmailService) SendEmail(message *EmailMessage) error {
	if err := es.validateMessage(message); err != nil {
		return fmt.Errorf("invalid email message: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= es.config.MaxRetries; attempt++ {
		if attempt > 0 {
			if es.logger != nil {
				es.logger.Printf("Retry attempt %d for email to %v", attempt, message.To)
			}
			time.Sleep(es.config.RetryDelay * time.Duration(attempt))
		}

		err := es.sendEmailAttempt(message)
		if err == nil {
			if es.logger != nil {
				es.logger.Printf("Email sent successfully to %v (attempt %d)", message.To, attempt+1)
			}
			return nil
		}

		lastErr = err
		if es.logger != nil {
			es.logger.Printf("Email send attempt %d failed: %v", attempt+1, err)
		}
	}

	return fmt.Errorf("failed to send email after %d attempts: %w", es.config.MaxRetries+1, lastErr)
}

// sendEmailAttempt performs a single email send attempt
func (es *EmailService) sendEmailAttempt(message *EmailMessage) error {
	client, err := es.createSMTPClient()
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if err := client.Auth(smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.SMTPHost)); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(es.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	allRecipients := append(message.To, message.CC...)
	allRecipients = append(allRecipients, message.BCC...)
	for _, recipient := range allRecipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email content
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to initialize data transfer: %w", err)
	}

	emailContent := es.buildEmailContent(message)
	if _, err := writer.Write([]byte(emailContent)); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close email data transfer: %w", err)
	}

	return nil
}

// SendSimpleEmail sends a simple text email
func (es *EmailService) SendSimpleEmail(to []string, subject, body string) error {
	message := &EmailMessage{
		To:       to,
		Subject:  subject,
		Body:     body,
		Priority: PriorityNormal,
	}
	return es.SendEmail(message)
}

// SendHTMLEmail sends an HTML email
func (es *EmailService) SendHTMLEmail(to []string, subject, htmlBody string) error {
	message := &EmailMessage{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		Priority: PriorityNormal,
	}
	return es.SendEmail(message)
}

// loadLogoAsBase64 loads the logo file and encodes it as base64
func (es *EmailService) loadLogoAsBase64() (string, error) {
	logoPath := filepath.Join("backend", "assets", "logo.png")

	// Check if logo file exists
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		// If logo doesn't exist, return empty string (template will handle gracefully)
		if es.logger != nil {
			es.logger.Printf("Logo file not found at %s, emails will be sent without logo", logoPath)
		}
		return "", nil
	}

	// Read the logo file
	logoData, err := os.ReadFile(logoPath)
	if err != nil {
		if es.logger != nil {
			es.logger.Printf("Failed to read logo file: %v", err)
		}
		return "", nil // Return empty string instead of error to allow emails to be sent
	}

	// Encode to base64
	logoBase64 := base64.StdEncoding.EncodeToString(logoData)
	return logoBase64, nil
}

// SendTemplatedEmail sends an email using a template
func (es *EmailService) SendTemplatedEmail(to []string, templateName string, data map[string]any) error {
	template, exists := es.templates[templateName]
	if !exists {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Ensure required data is present
	if data == nil {
		data = make(map[string]any)
	}
	data["ServerURL"] = FRONTEND
	data["Email"] = strings.Join(to, ", ")
	data["Date"] = time.Now().Format("January 2, 2006 at 3:04 PM")

	// Load and add logo as base64
	logoBase64, err := es.loadLogoAsBase64()
	// logoBase64 := (`image:data64`)
	if err != nil {
		// Log error but don't fail the email send
		if es.logger != nil {
			es.logger.Printf("Failed to load logo: %v", err)
		}
	}
	data["LogoBase64"] = logoBase64

	// Render templates
	subject, err := es.renderTemplate(template.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to render subject template: %w", err)
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))

	htmlBody, err := es.renderTemplate(template.HTMLContent, data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := es.renderTemplate(template.TextContent, data)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	message := &EmailMessage{
		To:       to,
		Subject:  subject,
		Body:     textBody,
		HTMLBody: htmlBody,
		Priority: PriorityNormal,
	}

	return es.SendEmail(message)
}

// SendWelcomeEmail sends a welcome email to new users
func (es *EmailService) SendWelcomeEmail(to, userName, token string) error {
	data := map[string]any{
		"UserName": userName,
		"Token":    token,
	}
	return es.SendTemplatedEmail([]string{to}, "welcome", data)
}

// SendEmailVerification sends an email verification email
func (es *EmailService) SendEmailVerification(to, userName, verificationToken string) error {
	data := map[string]any{
		"UserName":          userName,
		"VerificationToken": verificationToken,
	}
	return es.SendTemplatedEmail([]string{to}, "welcome", data)
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(to string, resetToken string) error {
	data := map[string]any{
		"ResetToken": resetToken,
	}
	return es.SendTemplatedEmail([]string{to}, "password_reset", data)
}

func (es *EmailService) SendTransactionApprovedEmail(to string, userName string, transaction models.Transaction) error {
	// Get dynamic data based on transaction type
	dynamicData := templates.GetApprovedTransactionData(string(transaction.TransactionType), transaction)

	data := map[string]any{
		"UserName":        userName,
		"Email":           to,
		"TransactionID":   transaction.TransactionID,
		"TransactionType": transaction.TransactionType,
		"Amount":          utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		"Date":            transaction.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		"ServerURL":       FRONTEND,
		"RecipientName":   transaction.TransactionDetails.RecipientName,
		"BankName":        transaction.TransactionDetails.BankName,
		"AccountNumber":   transaction.TransactionDetails.AccountNumber,
		"PhoneNumber":     transaction.TransactionDetails.PhoneNumber,
		"Network":         transaction.TransactionDetails.Network,
		"FromAmount":      utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		"ToAmount":        utils.FormatCurrency(transaction.TransactionDetails.ToAmount, transaction.TransactionDetails.ToCurrency),
		"FromCurrency":    transaction.TransactionDetails.FromCurrency,
		"ToCurrency":      transaction.TransactionDetails.ToCurrency,
		"ExchangeRate":    fmt.Sprintf("1 %s = %.4f %s", transaction.TransactionDetails.FromCurrency, transaction.TransactionDetails.ToAmount/transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.ToCurrency),
	}

	// Merge dynamic data
	for key, value := range dynamicData {
		data[key] = value
	}

	return es.SendTemplatedEmail([]string{to}, "transaction_approved", data)
}

func (es *EmailService) SendTransactionRejectedEmail(to string, userName string, transaction models.Transaction, reason string) error {
	// Get dynamic data based on transaction type
	dynamicData := templates.GetRejectedTransactionData(string(transaction.TransactionType), transaction)

	// Get user-friendly reason
	friendlyReason := templates.GetUserFriendlyRejectionReason(reason, string(transaction.TransactionType))

	data := map[string]any{
		"UserName":        userName,
		"Email":           to,
		"TransactionID":   transaction.TransactionID,
		"TransactionType": transaction.TransactionType,
		"Amount":          utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		"Date":            transaction.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		"Reason":          friendlyReason,
		"ServerURL":       FRONTEND,
		"RecipientName":   transaction.TransactionDetails.RecipientName,
		"BankName":        transaction.TransactionDetails.BankName,
		"AccountNumber":   transaction.TransactionDetails.AccountNumber,
		"PhoneNumber":     transaction.TransactionDetails.PhoneNumber,
		"Network":         transaction.TransactionDetails.Network,
		"FromAmount":      utils.FormatCurrency(transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.FromCurrency),
		"ToAmount":        utils.FormatCurrency(transaction.TransactionDetails.ToAmount, transaction.TransactionDetails.ToCurrency),
		"FromCurrency":    transaction.TransactionDetails.FromCurrency,
		"ToCurrency":      transaction.TransactionDetails.ToCurrency,
		"ExchangeRate":    fmt.Sprintf("1 %s = %.4f %s", transaction.TransactionDetails.FromCurrency, transaction.TransactionDetails.ToAmount/transaction.TransactionDetails.FromAmount, transaction.TransactionDetails.ToCurrency),
	}

	// Merge dynamic data
	for key, value := range dynamicData {
		data[key] = value
	}

	return es.SendTemplatedEmail([]string{to}, "transaction_rejected", data)
}

// renderTemplate renders a template with the given data
func (es *EmailService) renderTemplate(templateContent string, data map[string]any) (string, error) {
	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Send TwoFactorAuthenticationEmail sends a 2FA email
func (es *EmailService) SendTwoFactorAuthenticationEmail(to, userName, verificationCode string) error {
	data := map[string]any{
		"UserName":         userName,
		"VerificationCode": verificationCode,
	}
	return es.SendTemplatedEmail([]string{to}, "two_factor_authentication", data)
}

// validateMessage validates an email message
func (es *EmailService) validateMessage(message *EmailMessage) error {
	if len(message.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if message.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if message.Body == "" && message.HTMLBody == "" {
		return fmt.Errorf("either body or HTML body is required")
	}

	// Validate email addresses
	allEmails := append(message.To, message.CC...)
	allEmails = append(allEmails, message.BCC...)

	for _, email := range allEmails {
		if !IsValidEmail(email) {
			return fmt.Errorf("invalid email address: %s", email)
		}
	}

	return nil
}

// createSMTPClient creates and configures an SMTP client
func (es *EmailService) createSMTPClient() (*smtp.Client, error) {
	serverAddr := fmt.Sprintf("%s:%d", es.config.SMTPHost, es.config.SMTPPort)

	if es.config.UseTLS {
		// Connect with TLS
		tlsConfig := &tls.Config{
			InsecureSkipVerify: es.config.InsecureTLS,
			ServerName:         es.config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect with TLS: %w", err)
		}

		client, err := smtp.NewClient(conn, es.config.SMTPHost)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create SMTP client: %w", err)
		}

		return client, nil
	} else {
		// Connect without TLS (for testing or specific configurations)
		client, err := smtp.Dial(serverAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial SMTP server: %w", err)
		}

		// Try to enable STARTTLS if available
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: es.config.InsecureTLS,
				ServerName:         es.config.SMTPHost,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Quit()
				return nil, fmt.Errorf("failed to start TLS: %w", err)
			}
		}

		return client, nil
	}
}

// buildEmailContent constructs the email content with proper headers
func (es *EmailService) buildEmailContent(message *EmailMessage) string {
	var content strings.Builder

	// Headers
	content.WriteString(fmt.Sprintf("From: %s <%s>\r\n", es.config.FromName, es.config.FromEmail))
	content.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))

	if len(message.CC) > 0 {
		content.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.CC, ", ")))
	}

	content.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	content.WriteString("MIME-Version: 1.0\r\n")

	// Add priority header
	if message.Priority != PriorityNormal {
		switch message.Priority {
		case PriorityHigh:
			content.WriteString("X-Priority: 2\r\n")
			content.WriteString("Importance: High\r\n")
		case PriorityUrgent:
			content.WriteString("X-Priority: 1\r\n")
			content.WriteString("Importance: High\r\n")
		case PriorityLow:
			content.WriteString("X-Priority: 4\r\n")
			content.WriteString("Importance: Low\r\n")
		}
	}

	// Add custom headers
	for key, value := range message.Headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Add date header
	content.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	// Determine content type
	if message.HTMLBody != "" && message.Body != "" {
		// Mixed content (both HTML and plain text)
		boundary := fmt.Sprintf("boundary_%d", time.Now().Unix())
		content.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))

		// Plain text part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.Body)
		content.WriteString("\r\n\r\n")

		// HTML part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
		content.WriteString("\r\n\r\n")

		content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if message.HTMLBody != "" {
		// HTML only
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
	} else {
		// Plain text only
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.Body)
	}

	return content.String()
}

// TestConnection tests the SMTP connection
func (es *EmailService) TestConnection() error {
	client, err := es.createSMTPClient()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.SMTPHost)); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if es.logger != nil {
		es.logger.Println("SMTP connection test successful")
	}

	return nil
}

// AddTemplate adds a custom email template
func (es *EmailService) AddTemplate(template *EmailTemplate) {
	es.templates[template.Name] = template
}

// GetTemplate retrieves an email template by name
func (es *EmailService) GetTemplate(name string) (*EmailTemplate, bool) {
	template, exists := es.templates[name]
	return template, exists
}

// ListTemplates returns all available template names
func (es *EmailService) ListTemplates() []string {
	names := make([]string, 0, len(es.templates))
	for name := range es.templates {
		names = append(names, name)
	}
	return names
}

// IsValidEmail validates an email address using regex
func IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateEmailBatch validates multiple email addresses
func ValidateEmailBatch(emails []string) map[string]EmailValidationResult {
	results := make(map[string]EmailValidationResult)

	for _, email := range emails {
		if IsValidEmail(email) {
			results[email] = EmailValidationResult{IsValid: true}
		} else {
			results[email] = EmailValidationResult{
				IsValid: false,
				Error:   "Invalid email format",
			}
		}
	}

	return results
}

// Helper functions for environment variables

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEmailStats returns statistics about email operations
type EmailStats struct {
	TotalSent   int64
	TotalFailed int64
	LastSent    time.Time
	LastError   string
}

// This is a placeholder for email statistics tracking
// In a real implementation, you would store these in a database or cache
var emailStats = &EmailStats{}

func (es *EmailService) GetStats() *EmailStats {
	return emailStats
}

// UpdateStats updates email statistics (placeholder implementation)
// func (es *EmailService) updateStats(success bool, err error) {
// 	if success {
// 		emailStats.TotalSent++
// 		emailStats.LastSent = time.Now()
// 	} else {
// 		emailStats.TotalFailed++
// 		if err != nil {
// 			emailStats.LastError = err.Error()
// 		}
// 	}
// }

// Ensure EmailService implements the EmailSender interface
var _ interfaces.EmailSender = (*EmailService)(nil)

// InitializeGlobalEmailSender initializes the global email sender
func InitializeGlobalEmailSender() error {
	emailService, err := NewEmailServiceFromEnv()
	if err != nil {
		return fmt.Errorf("failed to initialize email service: %w", err)
	}

	interfaces.SetGlobalEmailSender(emailService)
	return nil
}
