package constants

import (
	"os"
	"strconv"
	"time"
)

// Environment types
const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

// Currency types
const (
	CurrencyNGN = "NGN"
	CurrencyGHS = "GHS"
)

// Transaction types
const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypeConversion = "conversion"
	TransactionTypeTransfer   = "transfer"
)

// Transaction status
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

// Transaction directions
const (
	DirectionNGNToGHS      = "NGN-GHS"
	DirectionGHSToNGN      = "GHS-NGN"
	DirectionDepositNGN    = "DEPOSIT-NGN"
	DirectionDepositGHS    = "DEPOSIT_GHS"
	DirectionWithdrawalNGN = "WITHDRAWAL_NGN"
	DirectionWithdrawalGHS = "WITHDRAWAL_GHS"
)

// Payment methods
const (
	PaymentMethodBank = "bank"
	PaymentMethodMomo = "momo"
)

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Notification types
const (
	NotificationTypeSuccess = "success"
	NotificationTypeInfo    = "info"
	NotificationTypeWarning = "warning"
	NotificationTypeError   = "error"
)

// Admin action types
const (
	AdminActionApproveTransaction = "approve_transaction"
	AdminActionRejectTransaction  = "reject_transaction"
	AdminActionBlockUser          = "block_user"
	AdminActionUnblockUser        = "unblock_user"
	AdminActionDeleteUser         = "delete_user"
	AdminActionUpdateRate         = "update_rate"
)

// Default values
const (
	DefaultPageSize      = 20
	MaxPageSize          = 100
	ConversionFeePercent = 0.02 // 2%
	DefaultExchangeRate  = 1.0
)

// JWT constants
const (
	JWTIssuer            = "jeanpay"
	JWTDefaultExpiration = 24 * time.Hour
	JWTRefreshExpiration = 7 * 24 * time.Hour
)

// Validation constants
const (
	MinPasswordLength = 6
	MaxPasswordLength = 100
	MinNameLength     = 2
	MaxNameLength     = 50
	MaxEmailLength    = 255
)

// File upload constants
const (
	MaxFileSize      = 10 * 1024 * 1024 // 10MB
	AllowedImageExts = "jpg,jpeg,png,gif"
	AllowedDocExts   = "pdf,doc,docx"
)

// Rate limiting
const (
	RateLimitAuth           = 5   // 5 requests per minute for auth endpoints
	RateLimitGeneral        = 100 // 100 requests per minute for general endpoints
	RateLimitWindow         = time.Minute
	RateLimitAdminEndpoints = 200 // 200 requests per minute for admin endpoints
)

// Configuration struct for environment variables
type Config struct {
	Environment       string
	Port              int
	DatabaseURL       string
	JWTSecret         string
	JWTExpiration     string
	PaystackSecretKey string
	PaystackPublicKey string
	MomoAPIKey        string
	MomoAPISecret     string
	EmailSMTPHost     string
	EmailSMTPPort     int
	EmailUsername     string
	EmailPassword     string
	RedisURL          string
	AllowedOrigins    []string
}

// GetEnv gets environment variable with fallback
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvInt gets environment variable as integer with fallback
func GetEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnvBool gets environment variable as boolean with fallback
func GetEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Environment:       GetEnv("GO_ENV", EnvDevelopment),
		Port:              GetEnvInt("PORT", 8080),
		DatabaseURL:       GetEnv("DATABASE_URL", "mongodb://localhost:27017/jeanpay"),
		JWTSecret:         GetEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiration:     GetEnv("JWT_EXPIRATION", "24h"),
		PaystackSecretKey: GetEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey: GetEnv("PAYSTACK_PUBLIC_KEY", ""),
		MomoAPIKey:        GetEnv("MOMO_API_KEY", ""),
		MomoAPISecret:     GetEnv("MOMO_API_SECRET", ""),
		EmailSMTPHost:     GetEnv("EMAIL_SMTP_HOST", "smtp.gmail.com"),
		EmailSMTPPort:     GetEnvInt("EMAIL_SMTP_PORT", 587),
		EmailUsername:     GetEnv("EMAIL_USERNAME", ""),
		EmailPassword:     GetEnv("EMAIL_PASSWORD", ""),
		RedisURL:          GetEnv("REDIS_URL", "redis://localhost:6379"),
		AllowedOrigins:    []string{GetEnv("ALLOWED_ORIGINS", "*")},
	}
}

// IsValidCurrency checks if currency is valid
func IsValidCurrency(currency string) bool {
	return currency == CurrencyNGN || currency == CurrencyGHS
}

// IsValidTransactionType checks if transaction type is valid
func IsValidTransactionType(txType string) bool {
	switch txType {
	case TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeConversion, TransactionTypeTransfer:
		return true
	default:
		return false
	}
}

// IsValidTransactionStatus checks if transaction status is valid
func IsValidTransactionStatus(status string) bool {
	switch status {
	case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed:
		return true
	default:
		return false
	}
}

// IsValidPaymentMethod checks if payment method is valid
func IsValidPaymentMethod(method string) bool {
	switch method {
	case PaymentMethodBank, PaymentMethodMomo:
		return true
	default:
		return false
	}
}

// GetSupportedCurrencies returns list of supported currencies
func GetSupportedCurrencies() []string {
	return []string{CurrencyNGN, CurrencyGHS}
}

// GetSupportedCountries returns list of supported countries
func GetSupportedCountries() []string {
	return []string{"NGN", "GHS", "ZAR", "TZA", "UGA", "RWA", "ZMB", "MWI", "BWA", "ZWE"}
}

// IsDevelopment checks if running in development mode
func IsDevelopment() bool {
	return GetEnv("GO_ENV", EnvDevelopment) == EnvDevelopment
}

// IsProduction checks if running in production mode
func IsProduction() bool {
	return GetEnv("GO_ENV", EnvDevelopment) == EnvProduction
}

// IsTest checks if running in test mode
func IsTest() bool {
	return GetEnv("GO_ENV", EnvDevelopment) == EnvTest
}
