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
	SeedDefaultData()
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
	)
}
