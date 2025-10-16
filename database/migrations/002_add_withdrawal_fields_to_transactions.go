package migrations

import (
	"gorm.io/gorm"
)

// Migration002AddWithdrawalFieldsToTransactions adds withdrawal_fee and fee_calculation fields to transactions table
func Migration002AddWithdrawalFieldsToTransactions(db *gorm.DB) error {
	// Add withdrawal_fee column if it doesn't exist
	if !db.Migrator().HasColumn("transactions", "withdrawal_fee") {
		if err := db.Migrator().AddColumn("transactions", "withdrawal_fee"); err != nil {
			return err
		}
	}

	// Add fee_calculation column if it doesn't exist
	if !db.Migrator().HasColumn("transactions", "fee_calculation") {
		if err := db.Migrator().AddColumn("transactions", "fee_calculation"); err != nil {
			return err
		}
	}

	// Set default values for existing transactions
	if err := db.Exec("UPDATE transactions SET withdrawal_fee = 0 WHERE withdrawal_fee IS NULL").Error; err != nil {
		return err
	}

	return nil
}
