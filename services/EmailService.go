package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/interfaces"
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
	SERVER = GetEnvOrDefault("SERVER_URL", "http://localhost:8080")

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
		SMTPHost:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvIntOrDefault("SMTP_PORT", 587),
		Username:     os.Getenv("SMTP_USERNAME"),
		Password:     os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
		FromName:     getEnvOrDefault("FROM_NAME", "JeanPay"),
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
	es.templates["welcome"] = &EmailTemplate{
		Name:    "welcome",
		Subject: "Welcome to JeanPay!",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to JeanPay</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 15px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to JeanPay!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.UserName}}!</h2>
            <p>Thank you for joining our platform. We're excited to have you on board!</p>
            <p>You can now:</p>
            <ul>
                <li>Make secure payments</li>
                <li>Manage your account</li>
                <li>Track your transactions</li>
            </ul>
            <p>To activate your account, please click the button below:</p>
            <a href="{{.ServerURL}}/activate?token={{.Token}}" class="button">Activate Account</a>
            <p>If you have any questions, feel free to contact our support team.</p>
        </div>
        <div class="footer">
            <p>Best regards,<br>The JeanPay Team</p>
            <p>This email was sent to {{.Email}}. If you did not sign up for JeanPay, please ignore this email.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `Welcome to JeanPay, {{.UserName}}!

Thank you for joining our platform. We're excited to have you on board!

You can now:
- Make secure payments
- Manage your account
- Track your transactions

To activate your account, please visit: {{.ServerURL}}/activate?token={{.Token}}

If you have any questions, feel free to contact our support team.

Best regards,
The JeanPay Team`,
	}

	es.templates["password_reset"] = &EmailTemplate{
		Name:    "password_reset",
		Subject: "Password Reset Request - JeanPay",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 5px; margin: 15px 0; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>You have requested to reset your password for your JeanPay account.</p>
            <p>Please click the button below to reset your password:</p>
            <a href="{{.ServerURL}}/reset-password?token={{.ResetToken}}" class="button">Reset Password</a>
            <div class="warning">
                <strong>⚠️ This link will expire in 1 hour.</strong>
            </div>
            <p>If you did not request this password reset, please ignore this email and your password will remain unchanged.</p>
        </div>
        <div class="footer">
            <p>Best regards,<br>The JeanPay Team</p>
            <p>This email was sent to {{.Email}}.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `Password Reset Request - JeanPay

You have requested to reset your password for your JeanPay account.

Please visit the following link to reset your password:
{{.ServerURL}}/reset-password?token={{.ResetToken}}

⚠️ This link will expire in 1 hour.

If you did not request this password reset, please ignore this email and your password will remain unchanged.

Best regards,
The JeanPay Team`,
	}

	es.templates["transaction_notification"] = &EmailTemplate{
		Name:    "transaction_notification",
		Subject: "Transaction {{.TransactionType}} - JeanPay",
		HTMLContent: `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Transaction {{.TransactionType}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .transaction-details { background-color: white; border-left: 4px solid #28a745; padding: 15px; margin: 15px 0; }
        .amount { font-size: 24px; font-weight: bold; color: #28a745; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Transaction {{.TransactionType}}</h1>
        </div>
        <div class="content">
            <p>Your transaction has been processed successfully.</p>
            <div class="transaction-details">
                <p><strong>Transaction Type:</strong> {{.TransactionType}}</p>
                <p><strong>Amount:</strong> <span class="amount">{{.Amount}}</span></p>
                <p><strong>Transaction ID:</strong> {{.TransactionID}}</p>
                <p><strong>Date:</strong> {{.Date}}</p>
            </div>
            <p>You can view your complete transaction history in your JeanPay account dashboard.</p>
        </div>
        <div class="footer">
            <p>Best regards,<br>The JeanPay Team</p>
            <p>This email was sent to {{.Email}}.</p>
        </div>
    </div>
</body>
</html>`,
		TextContent: `Transaction {{.TransactionType}} - JeanPay

Your transaction has been processed successfully.

Transaction Details:
- Type: {{.TransactionType}}
- Amount: {{.Amount}}
- Transaction ID: {{.TransactionID}}
- Date: {{.Date}}

You can view your complete transaction history in your JeanPay account dashboard.

Best regards,
The JeanPay Team`,
	}

	es.templates["two_factor_authentication"] = &EmailTemplate{
		Name:    "two_factor_authentication",
		Subject: "Two-Factor Authentication - JeanPay",
		HTMLContent: `
<!DOCTYPE html>
		<html>
		<head>
						<meta charset="UTF-8">
						<meta name="viewport" content="width=device-width, initial-scale=1.0">
						<title>Two-Factor Authentication</title>
						<style>
										body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
										.container { max-width: 600px; margin: 0 auto; padding: 20px; }
										.header { background-color: #17a2b8; color: white; padding: 20px; text-align: center; }
										.content { padding: 20px; background-color: #f9f9f9; }
										.button { display: inline-block; padding: 12px 24px; background-color: #17a2b8; color: white; text-decoration: none; border-radius: 5px; margin: 15px 0; }
										.footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
										.code-box { font-size: 24px; font-weight: bold; background: #e9ecef; padding: 10px 20px; border-radius: 5px; display: inline-block; margin: 15px 0; letter-spacing: 2px; }
						</style>
		</head>
		<body>
						<div class="container">
										<div class="header">
														<h1>Two-Factor Authentication</h1>
										</div>
										<div class="content">
														<h2>Hello {{.UserName}}!</h2>
														<p> Your login verification codew:</p>
	<div class="code-box" style="font-size: 24px; font-weight: bold; background: #e9ecef; padding: 10px 20px; border-radius: 5px; display: inline-block; margin: 15px 0; letter-spacing: 2px;">
															{{.VerificationCode}}
														</div>
														<p>This code will expire in 10 minutes.</p>
														<p>If you did not request this verification, please ignore this email.</p>
										</div>
										<div class="footer">
														<p>Best regards,<br>The JeanPay Team</p>
														<p>This email was sent to {{.Email}}.</p>
										</div>
						</div>
		</body>
		</html>
		`,
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
	data["ServerURL"] = SERVER
	data["Email"] = strings.Join(to, ", ")
	data["Date"] = time.Now().Format("January 2, 2006 at 3:04 PM")

	// Render templates
	subject, err := es.renderTemplate(template.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to render subject template: %w", err)
	}

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
func (es *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	data := map[string]any{
		"ResetToken": resetToken,
	}
	return es.SendTemplatedEmail([]string{to}, "password_reset", data)
}

// SendTransactionNotification sends a transaction notification email
func (es *EmailService) SendTransactionNotification(to, transactionType, amount, transactionID string) error {
	data := map[string]any{
		"TransactionType": transactionType,
		"Amount":          amount,
		"TransactionID":   transactionID,
	}
	return es.SendTemplatedEmail([]string{to}, "transaction_notification", data)
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
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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
