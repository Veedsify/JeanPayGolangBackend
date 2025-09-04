package database

import (
	"fmt"
	"regexp"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var HOST string = libs.GetEnvOrDefault("DB_HOST", "localhost")
var USER string = libs.GetEnvOrDefault("DB_USER", "postgres")
var PASSWORD string = libs.GetEnvOrDefault("DB_PASSWORD", "1234")
var NAME string = libs.GetEnvOrDefault("DB_NAME", "jeanpay")
var PORT string = libs.GetEnvOrDefault("DB_PORT", "5432")
var AdminEmail = libs.GetEnvOrDefault("ADMIN_DEFAULT_EMAIL", "admin@jeanpay.com")
var AdminPassword = libs.GetEnvOrDefault("ADMIN_DEFAULT_PASSWORD", "admin123")

// validateAndFixAdminEmail ensures the admin email is valid
func validateAndFixAdminEmail() string {
	email := AdminEmail

	// If the email is the literal environment variable name or empty, use default
	if email == "" || email == "ADMIN_DEFAULT_EMAIL" || !isValidEmail(email) {
		fmt.Printf("Warning: Invalid or missing admin email '%s', using default: admin@jeanpay.com\n", email)
		return "admin@jeanpay.com"
	}

	return email
}

// isValidEmail validates an email address using regex
func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func InitDB() {
	dsn := "host=" + HOST + " user=" + USER + " password=" + PASSWORD + " dbname=" + NAME + " port=" + PORT + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB = db
	autoMigrate(db)
}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&models.User{},
		&models.AdminLog{},
		&models.ExchangeRate{},
		&models.Conversions{},
		&models.Notification{},
		&models.Rate{},
		&models.Transaction{},
		&models.TransactionDetails{},
		&models.Wallet{},
		&models.WebhookEvent{},
		&models.Setting{},
		&models.Activity{},
		&models.WithdrawMethod{},
		&models.SavedRecipient{},
		&models.PlatformSetting{},
		&models.PaymentAccount{},
		&models.Contact{},
	)

	// Seed admin user if it doesn't exist
	var count int64
	db.Model(&models.User{}).Where("email = ?", AdminEmail).Count(&count)
	hashedPassword, err := libs.HashPassword(AdminPassword)
	if err != nil {
		panic("failed to hash password")
	}
	if count == 0 {
		// Validate and fix admin email
		validEmail := validateAndFixAdminEmail()

		db.Create(&models.User{
			FirstName:          "Admin",
			LastName:           "User",
			Email:              validEmail,
			Username:           "admin",
			Password:           hashedPassword,
			IsAdmin:            true,
			IsVerified:         true,
			IsTwoFactorEnabled: false,
			ProfilePicture:     "/images/defaults/user.jpg",
			PhoneNumber:        "08012345678",
			Country:            models.Nigeria,
			Setting: models.Setting{
				DefaultCurrency: models.DefaultCurrency("NGN"),
				FeesBreakdown:   true,
			},
		})
		fmt.Printf("Admin user created with email: %s\n", validEmail)
	}
}
