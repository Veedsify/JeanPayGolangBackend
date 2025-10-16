package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// RunMigrations executes all pending migrations
func RunMigrations(db *gorm.DB) error {
	fmt.Println("Running database migrations...")

	// Run migration 001
	if err := Migration001AddPhoneFieldsToUsers(db); err != nil {
		return fmt.Errorf("failed to run migration 001: %w", err)
	}

	// Run migration 002
	if err := Migration002AddWithdrawalFieldsToTransactions(db); err != nil {
		return fmt.Errorf("failed to run migration 002: %w", err)
	}

	fmt.Println("All migrations completed successfully")
	return nil
}
