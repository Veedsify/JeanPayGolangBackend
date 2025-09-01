package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// GetPaymentAccounts retrieves payment accounts with optional filters and pagination
func GetPaymentAccounts(currency, accountType string, isActive *bool, search string, page, limit int) ([]models.PaymentAccount, int64, error) {
	var accounts []models.PaymentAccount
	var total int64

	query := database.DB

	// Apply filters
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(account_name) LIKE ? OR LOWER(bank_name) LIKE ? OR LOWER(network) LIKE ? OR account_number LIKE ? OR phone_number LIKE ?",
			searchTerm, searchTerm, searchTerm, search, search)
	}

	// Count total records
	if err := query.Model(&models.PaymentAccount{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment accounts: %v", err)
	}

	// Apply pagination and fetch records
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch payment accounts: %v", err)
	}

	return accounts, total, nil
}

// CreatePaymentAccount creates a new payment account
func CreatePaymentAccount(params types.CreatePaymentAccountRequest, createdBy uint) (*models.PaymentAccount, error) {
	// Validate account type specific fields
	if err := validatePaymentAccountFields(params); err != nil {
		return nil, err
	}

	account := &models.PaymentAccount{
		Currency:      params.Currency,
		AccountType:   models.AccountType(params.AccountType),
		AccountName:   params.AccountName,
		AccountNumber: params.AccountNumber,
		BankName:      params.BankName,
		BankCode:      params.BankCode,
		PhoneNumber:   params.PhoneNumber,
		Network:       params.Network,
		IsActive:      true,
		CreatedBy:     createdBy,
		Notes:         params.Notes,
	}

	if err := database.DB.Create(account).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment account: %v", err)
	}

	return account, nil
}

// UpdatePaymentAccount updates an existing payment account
func UpdatePaymentAccount(id uint, params types.UpdatePaymentAccountRequest) (*models.PaymentAccount, error) {
	var account models.PaymentAccount

	// Check if account exists
	if err := database.DB.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment account not found")
		}
		return nil, fmt.Errorf("failed to fetch payment account: %v", err)
	}

	// Update fields if provided
	updates := make(map[string]interface{})

	if params.Currency != nil {
		updates["currency"] = *params.Currency
	}
	if params.AccountType != nil {
		updates["account_type"] = *params.AccountType
	}
	if params.AccountName != nil {
		updates["account_name"] = *params.AccountName
	}
	if params.AccountNumber != nil {
		updates["account_number"] = *params.AccountNumber
	}
	if params.BankName != nil {
		updates["bank_name"] = *params.BankName
	}
	if params.BankCode != nil {
		updates["bank_code"] = *params.BankCode
	}
	if params.PhoneNumber != nil {
		updates["phone_number"] = *params.PhoneNumber
	}
	if params.Network != nil {
		updates["network"] = *params.Network
	}
	if params.Notes != nil {
		updates["notes"] = *params.Notes
	}
	if params.IsActive != nil {
		updates["is_active"] = *params.IsActive
	}

	if err := database.DB.Model(&account).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update payment account: %v", err)
	}

	return &account, nil
}

// DeletePaymentAccount deletes a payment account
func DeletePaymentAccount(id uint) error {
	var account models.PaymentAccount

	// Check if account exists
	if err := database.DB.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("payment account not found")
		}
		return fmt.Errorf("failed to fetch payment account: %v", err)
	}

	if err := database.DB.Delete(&account).Error; err != nil {
		return fmt.Errorf("failed to delete payment account: %v", err)
	}

	return nil
}

// TogglePaymentAccountStatus toggles the active status of a payment account
func TogglePaymentAccountStatus(id uint) (*models.PaymentAccount, error) {
	var account models.PaymentAccount

	// Check if account exists
	if err := database.DB.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment account not found")
		}
		return nil, fmt.Errorf("failed to fetch payment account: %v", err)
	}

	// Toggle status
	account.IsActive = !account.IsActive

	if err := database.DB.Save(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to update payment account status: %v", err)
	}

	return &account, nil
}

// GetPaymentAccountByID retrieves a single payment account by ID
func GetPaymentAccountByID(id uint) (*models.PaymentAccount, error) {
	var account models.PaymentAccount

	if err := database.DB.First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment account not found")
		}
		return nil, fmt.Errorf("failed to fetch payment account: %v", err)
	}

	return &account, nil
}

// GetActivePaymentAccounts retrieves active payment accounts for topup functionality
func GetActivePaymentAccounts(currency, accountType string) ([]models.PaymentAccount, error) {
	var accounts []models.PaymentAccount

	query := database.DB.Where("is_active = ?", true)

	if currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}

	if err := query.Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active payment accounts: %v", err)
	}

	return accounts, nil
}

// validatePaymentAccountFields validates that required fields are provided based on account type
func validatePaymentAccountFields(params types.CreatePaymentAccountRequest) error {
	if params.AccountType == "bank" {
		if params.AccountNumber == "" {
			return errors.New("account number is required for bank accounts")
		}
		if params.BankName == "" {
			return errors.New("bank name is required for bank accounts")
		}
		if params.BankCode == "" {
			return errors.New("bank code is required for bank accounts")
		}
	} else if params.AccountType == "momo" {
		if params.PhoneNumber == "" {
			return errors.New("phone number is required for mobile money accounts")
		}
		if params.Network == "" {
			return errors.New("network is required for mobile money accounts")
		}
	}
	return nil
}

// GetPaymentAccountForTopup retrieves the appropriate payment account for topup based on currency and payment type
func GetPaymentAccountForTopup(currency, paymentType string) (*models.PaymentAccount, error) {
	var account models.PaymentAccount

	query := database.DB.Where("is_active = ? AND currency = ? AND account_type = ?", true, currency, paymentType)

	if err := query.First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active payment account found for %s %s", currency, paymentType)
		}
		return nil, fmt.Errorf("failed to fetch payment account: %v", err)
	}

	return &account, nil
}
