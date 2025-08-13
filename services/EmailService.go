package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// EmailConfig holds the SMTP server configuration
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromEmail   string
	FromName    string
	UseTLS      bool
	InsecureTLS bool
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
}

// EmailAttachment represents a file attachment
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// EmailService handles email operations
type EmailService struct {
	config *EmailConfig
}

func GetEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var SERVER = GetEnvOrDefault("SERVER_URL", "http://localhost:8080")

// NewEmailService creates a new email service instance
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// NewEmailServiceFromEnv creates email service from environment variables
func NewEmailServiceFromEnv() (*EmailService, error) {
	config := &EmailConfig{
		SMTPHost:    getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:    getEnvIntOrDefault("SMTP_PORT", 587),
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		FromEmail:   os.Getenv("FROM_EMAIL"),
		FromName:    getEnvOrDefault("FROM_NAME", "JeanPay"),
		UseTLS:      getEnvBoolOrDefault("SMTP_USE_TLS", true),
		InsecureTLS: getEnvBoolOrDefault("SMTP_INSECURE_TLS", false),
	}

	if config.Username == "" || config.Password == "" || config.FromEmail == "" {
		return nil, fmt.Errorf("missing required email configuration: SMTP_USERNAME, SMTP_PASSWORD, and FROM_EMAIL must be set")
	}

	return NewEmailService(config), nil
}

// SendEmail sends an email message
func (es *EmailService) SendEmail(message *EmailMessage) error {
	if len(message.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Create SMTP client
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
		To:      to,
		Subject: subject,
		Body:    body,
	}
	return es.SendEmail(message)
}

// SendHTMLEmail sends an HTML email
func (es *EmailService) SendHTMLEmail(to []string, subject, htmlBody string) error {
	message := &EmailMessage{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
	}
	err := es.SendEmail(message)
	if err != nil {
		return err
	}
	fmt.Printf("Email sent successfully to %v", to)
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (es *EmailService) SendWelcomeEmail(to, userName string, token string) error {
	subject := "Welcome to JeanPay!"
	htmlBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to JeanPay, %s!</h2>
			<p>Thank you for joining our platform. We're excited to have you on board!</p>
			<p>You can now:</p>
			<ul>
				<li>Make secure payments</li>
				<li>Manage your account</li>
				<li>Track your transactions</li>
			</ul>
			<div style="padding: 10px; margin: 10px 0; font-family: monospace; font-size: 16px;">
				<a style="color: #007bff; text-decoration: none; padding: 5px 10px; border-radius: 5px; background-color: #007bff; color: #fff;" href="%s/reset-password?token=%s">Activate Account</a>
			</div>
			<p>If you have any questions, feel free to contact our support team.</p>
			<br>
			<p>Best regards,<br>The JeanPay Team</p>
		</body>
		</html>
	`, userName, SERVER, token)

	return es.SendHTMLEmail([]string{to}, subject, htmlBody)
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	subject := "Password Reset Request - JeanPay"
	htmlBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>You have requested to reset your password for your JeanPay account.</p>
			<p>Please use the following token to reset your password:</p>
			<div style="padding: 10px; margin: 10px 0; font-family: monospace; font-size: 16px;">
				<a style="color: #007bff; text-decoration: none; padding: 5px 10px; border-radius: 5px; background-color: #007bff; color: #fff;" href="%s/reset-password?token=%s">Reset Password</a>
			</div>
			<p><strong>This token will expire in 1 hour.</strong></p>
			<p>If you did not request this password reset, please ignore this email.</p>
			<br>
			<p>Best regards,<br>The JeanPay Team</p>
		</body>
		</html>
	`, SERVER, resetToken)

	return es.SendHTMLEmail([]string{to}, subject, htmlBody)
}

// SendTransactionNotification sends a transaction notification email
func (es *EmailService) SendTransactionNotification(to, transactionType, amount, transactionID string) error {
	subject := fmt.Sprintf("Transaction %s - JeanPay", transactionType)
	htmlBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>Transaction %s</h2>
			<p>Your transaction has been processed successfully.</p>
			<div style="background-color: #f9f9f9; padding: 15px; margin: 15px 0; border-left: 4px solid #007bff;">
				<p><strong>Transaction Type:</strong> %s</p>
				<p><strong>Amount:</strong> %s</p>
				<p><strong>Transaction ID:</strong> %s</p>
			</div>
			<p>You can view your transaction history in your JeanPay account.</p>
			<br>
			<p>Best regards,<br>The JeanPay Team</p>
		</body>
		</html>
	`, transactionType, transactionType, amount, transactionID)

	return es.SendHTMLEmail([]string{to}, subject, htmlBody)
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

	// Determine content type
	if message.HTMLBody != "" && message.Body != "" {
		// Mixed content (both HTML and plain text)
		boundary := "boundary123456789"
		content.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))

		// Plain text part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		content.WriteString(message.Body)
		content.WriteString("\r\n\r\n")

		// HTML part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		content.WriteString(message.HTMLBody)
		content.WriteString("\r\n\r\n")

		content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if message.HTMLBody != "" {
		// HTML only
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		content.WriteString(message.HTMLBody)
	} else {
		// Plain text only
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
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

	return nil
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
