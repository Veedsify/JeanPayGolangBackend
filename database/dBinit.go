package database

import (
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
	)

	// Seed admin user if it doesn't exist
	var count int64
	db.Model(&models.User{}).Where("email = ?", "admin@jeanpay.africa").Count(&count)
	hashedPassword, err := libs.HashPassword("password")
	if err != nil {
		panic("failed to hash password")
	}
	if count == 0 {
		db.Create(&models.User{
			FirstName:          "Admin",
			LastName:           "User",
			Email:              "admin@jeanpay.africa",
			Username:           "admin",
			Password:           hashedPassword, // 🔴 don’t store plaintext passwords!
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
	}
}
