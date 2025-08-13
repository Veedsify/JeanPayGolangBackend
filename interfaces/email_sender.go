package interfaces

// EmailSender defines the interface for sending emails
type EmailSender interface {
	// SendWelcomeEmail sends a welcome email to a new user
	SendWelcomeEmail(to, userName, token string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(to, resetToken string) error

	// SendTransactionNotification sends a transaction notification email
	SendTransactionNotification(to, transactionType, amount, transactionID string) error

	// SendEmailVerification sends an email verification email
	SendEmailVerification(to, userName, verificationToken string) error

	// SendTemplatedEmail sends an email using a template
	SendTemplatedEmail(to []string, templateName string, data map[string]interface{}) error

	// SendSimpleEmail sends a simple text email
	SendSimpleEmail(to []string, subject, body string) error

	// SendHTMLEmail sends an HTML email
	SendHTMLEmail(to []string, subject, htmlBody string) error
}

// EmailJobHandler defines the interface for handling email jobs
type EmailJobHandler interface {
	// SetEmailSender sets the email sender implementation
	SetEmailSender(sender EmailSender)

	// GetEmailSender returns the current email sender
	GetEmailSender() EmailSender
}

// Global email sender instance that can be set at runtime
var GlobalEmailSender EmailSender

// SetGlobalEmailSender sets the global email sender implementation
func SetGlobalEmailSender(sender EmailSender) {
	GlobalEmailSender = sender
}

// GetGlobalEmailSender returns the global email sender
func GetGlobalEmailSender() EmailSender {
	return GlobalEmailSender
}
