package migrations

import (
	"gorm.io/gorm"
)

// Migration001AddPhoneFieldsToUsers adds phone_number, country_code, and phone_verified fields to users table
func Migration001AddPhoneFieldsToUsers(db *gorm.DB) error {
	// Add country_code column if it doesn't exist
	if !db.Migrator().HasColumn("users", "country_code") {
		if err := db.Migrator().AddColumn("users", "country_code"); err != nil {
			return err
		}
	}

	// Add phone_verified column if it doesn't exist
	if !db.Migrator().HasColumn("users", "phone_verified") {
		if err := db.Migrator().AddColumn("users", "phone_verified"); err != nil {
			return err
		}
	}

	// Modify phone_number column to be NOT NULL if it exists
	if db.Migrator().HasColumn("users", "phone_number") {
		// Update existing NULL phone_number values to empty string before making it NOT NULL
		if err := db.Exec("UPDATE users SET phone_number = '' WHERE phone_number IS NULL").Error; err != nil {
			return err
		}
	}

	// Set default values for existing users
	if err := db.Exec("UPDATE users SET country_code = '+234' WHERE country_code IS NULL OR country_code = ''").Error; err != nil {
		return err
	}

	if err := db.Exec("UPDATE users SET phone_verified = false WHERE phone_verified IS NULL").Error; err != nil {
		return err
	}

	return nil
}
