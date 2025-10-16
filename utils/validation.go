package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/Veedsify/JeanPayGoBackend/constants"
)

// EmailRegex is a regex pattern for email validation
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// PhoneRegex is a regex pattern for phone number validation
var PhoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, ", ")
}

// Add adds a new validation error
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > constants.MaxEmailLength {
		return fmt.Errorf("email is too long (max %d characters)", constants.MaxEmailLength)
	}

	if !EmailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < constants.MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", constants.MinPasswordLength)
	}

	if len(password) > constants.MaxPasswordLength {
		return fmt.Errorf("password is too long (max %d characters)", constants.MaxPasswordLength)
	}

	// Check for at least one letter and one number
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	return nil
}

// ValidateName validates user names (first name, last name)
func ValidateName(name, fieldName string) error {
	if name == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	if len(name) < constants.MinNameLength {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, constants.MinNameLength)
	}

	if len(name) > constants.MaxNameLength {
		return fmt.Errorf("%s is too long (max %d characters)", fieldName, constants.MaxNameLength)
	}

	// Check if name contains only letters, spaces, hyphens, and apostrophes
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("%s can only contain letters, spaces, hyphens, and apostrophes", fieldName)
	}

	return nil
}

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // Phone number is optional
	}

	// Remove any spaces or dashes
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	if !PhoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// ValidatePhoneNumberRequired validates phone number format (required for registration)
func ValidatePhoneNumberRequired(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number is required")
	}

	// Remove any spaces or dashes
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	if !PhoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone number format")
	}

	// Check minimum length (should be at least 10 digits after country code)
	if len(cleanPhone) < 10 {
		return fmt.Errorf("phone number is too short")
	}

	if len(cleanPhone) > 15 {
		return fmt.Errorf("phone number is too long")
	}

	return nil
}

// ValidateCountryCode validates country code for supported countries
func ValidateCountryCode(countryCode string) error {
	if countryCode == "" {
		return fmt.Errorf("country code is required")
	}

	if !constants.IsValidCountryCode(countryCode) {
		supportedCodes := constants.GetSupportedCountryCodes()
		return fmt.Errorf("unsupported country code. Supported codes: %s", strings.Join(supportedCodes, ", "))
	}

	return nil
}

// ValidatePhoneWithCountryCode validates phone number with country code
func ValidatePhoneWithCountryCode(phone, countryCode string) error {
	if err := ValidatePhoneNumberRequired(phone); err != nil {
		return err
	}

	if err := ValidateCountryCode(countryCode); err != nil {
		return err
	}

	// Normalize phone number by removing country code if it's included
	cleanPhone := strings.TrimSpace(phone)

	// Remove country code from phone if it starts with it
	if strings.HasPrefix(cleanPhone, countryCode) {
		cleanPhone = strings.TrimPrefix(cleanPhone, countryCode)
		cleanPhone = strings.TrimLeft(cleanPhone, " -")
	}

	// Remove leading zero if present (common in local format)
	cleanPhone = strings.TrimPrefix(cleanPhone, "0")

	// Validate based on country code
	switch countryCode {
	case constants.CountryCodeNigeria: // Nigeria
		if len(cleanPhone) != 10 {
			return fmt.Errorf("Nigerian phone number must be 10 digits after country code")
		}
		// Nigerian mobile numbers start with 7, 8, or 9
		if !strings.HasPrefix(cleanPhone, "7") && !strings.HasPrefix(cleanPhone, "8") && !strings.HasPrefix(cleanPhone, "9") {
			return fmt.Errorf("Nigerian mobile number must start with 7, 8, or 9")
		}
	case constants.CountryCodeGhana: // Ghana
		if len(cleanPhone) != 9 {
			return fmt.Errorf("Ghanaian phone number must be 9 digits after country code")
		}
		// Ghanaian mobile numbers typically start with 2, 5, or 0
		if !strings.HasPrefix(cleanPhone, "2") && !strings.HasPrefix(cleanPhone, "5") && !strings.HasPrefix(cleanPhone, "0") {
			return fmt.Errorf("Ghanaian mobile number format is invalid")
		}
	}

	return nil
}

// NormalizePhoneNumber normalizes phone number format
func NormalizePhoneNumber(phone, countryCode string) string {
	cleanPhone := strings.TrimSpace(phone)

	// Remove country code from phone if it's included
	if strings.HasPrefix(cleanPhone, countryCode) {
		cleanPhone = strings.TrimPrefix(cleanPhone, countryCode)
		cleanPhone = strings.TrimLeft(cleanPhone, " -")
	}

	// Remove leading zero if present
	cleanPhone = strings.TrimPrefix(cleanPhone, "0")

	// Return normalized format without country code (will be stored separately)
	return cleanPhone
}

// ValidateCurrency validates currency code
func ValidateCurrency(currency string) error {
	if currency == "" {
		return fmt.Errorf("currency is required")
	}

	if !constants.IsValidCurrency(currency) {
		return fmt.Errorf("invalid currency. Supported currencies: %s", strings.Join(constants.GetSupportedCurrencies(), ", "))
	}

	return nil
}

// ValidateAmount validates monetary amounts
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if amount > 999999999.99 {
		return fmt.Errorf("amount is too large")
	}

	return nil
}

// ValidateTransactionType validates transaction type
func ValidateTransactionType(txType string) error {
	if txType == "" {
		return fmt.Errorf("transaction type is required")
	}

	if !constants.IsValidTransactionType(txType) {
		return fmt.Errorf("invalid transaction type")
	}

	return nil
}

// ValidateTransactionStatus validates transaction status
func ValidateTransactionStatus(status string) error {
	if status == "" {
		return fmt.Errorf("transaction status is required")
	}

	if !constants.IsValidTransactionStatus(status) {
		return fmt.Errorf("invalid transaction status")
	}

	return nil
}

// ValidatePaymentMethod validates payment method
func ValidatePaymentMethod(method string) error {
	if method == "" {
		return fmt.Errorf("payment method is required")
	}

	if !constants.IsValidPaymentMethod(method) {
		return fmt.Errorf("invalid payment method")
	}

	return nil
}

// ValidateCountry validates country code
func ValidateCountry(country string) error {
	if country == "" {
		return fmt.Errorf("country is required")
	}

	supportedCountries := constants.GetSupportedCountries()
	for _, supported := range supportedCountries {
		if country == supported {
			return nil
		}
	}

	return fmt.Errorf("invalid country. Supported countries: %s", strings.Join(supportedCountries, ", "))
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) (int, int, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = constants.DefaultPageSize
	}

	if limit > constants.MaxPageSize {
		limit = constants.MaxPageSize
	}

	return page, limit, nil
}

// ValidateUserRegistration validates user registration data
func ValidateUserRegistration(firstName, lastName, email, password, phoneNumber, country string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateName(firstName, "first name"); err != nil {
		errors.Add("firstName", err.Error())
	}

	if err := ValidateName(lastName, "last name"); err != nil {
		errors.Add("lastName", err.Error())
	}

	if err := ValidateEmail(email); err != nil {
		errors.Add("email", err.Error())
	}

	if err := ValidatePassword(password); err != nil {
		errors.Add("password", err.Error())
	}

	if err := ValidatePhoneNumber(phoneNumber); err != nil {
		errors.Add("phoneNumber", err.Error())
	}

	if err := ValidateCountry(country); err != nil {
		errors.Add("country", err.Error())
	}

	return errors
}

// ValidateUserRegistrationWithPhone validates user registration data with required phone and country code
func ValidateUserRegistrationWithPhone(firstName, lastName, email, password, phoneNumber, countryCode, country string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateName(firstName, "first name"); err != nil {
		errors.Add("firstName", err.Error())
	}

	if err := ValidateName(lastName, "last name"); err != nil {
		errors.Add("lastName", err.Error())
	}

	if err := ValidateEmail(email); err != nil {
		errors.Add("email", err.Error())
	}

	if err := ValidatePassword(password); err != nil {
		errors.Add("password", err.Error())
	}

	if err := ValidatePhoneWithCountryCode(phoneNumber, countryCode); err != nil {
		errors.Add("phoneNumber", err.Error())
	}

	if err := ValidateCountry(country); err != nil {
		errors.Add("country", err.Error())
	}

	return errors
}

// ValidateUserLogin validates user login data
func ValidateUserLogin(email, password string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateEmail(email); err != nil {
		errors.Add("email", err.Error())
	}

	if password == "" {
		errors.Add("password", "password is required")
	}

	return errors
}

// ValidateConversionRequest validates currency conversion request
func ValidateConversionRequest(fromCurrency, toCurrency string, amount float64) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateCurrency(fromCurrency); err != nil {
		errors.Add("fromCurrency", err.Error())
	}

	if err := ValidateCurrency(toCurrency); err != nil {
		errors.Add("toCurrency", err.Error())
	}

	// Use the new currency pair validation
	if err := ValidateCurrencyPair(fromCurrency, toCurrency); err != nil {
		errors.Add("currencyPair", err.Error())
	}

	if err := ValidateAmount(amount); err != nil {
		errors.Add("amount", err.Error())
	}

	return errors
}

// ValidateConversionRequestWithCountry validates currency conversion request with country validation for XOF
func ValidateConversionRequestWithCountry(fromCurrency, toCurrency string, amount float64, countryCode string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateCurrency(fromCurrency); err != nil {
		errors.Add("fromCurrency", err.Error())
	}

	if err := ValidateCurrency(toCurrency); err != nil {
		errors.Add("toCurrency", err.Error())
	}

	// Use the new currency pair validation
	if err := ValidateCurrencyPair(fromCurrency, toCurrency); err != nil {
		errors.Add("currencyPair", err.Error())
	}

	// Validate XOF country restrictions
	if err := ValidateXOFCountry(fromCurrency, countryCode); err != nil {
		errors.Add("fromCurrency", err.Error())
	}

	if err := ValidateXOFCountry(toCurrency, countryCode); err != nil {
		errors.Add("toCurrency", err.Error())
	}

	if err := ValidateAmount(amount); err != nil {
		errors.Add("amount", err.Error())
	}

	return errors
}

// ValidateWalletTopUp validates wallet top-up request
func ValidateWalletTopUp(amount float64, currency, paymentMethod string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateAmount(amount); err != nil {
		errors.Add("amount", err.Error())
	}

	if err := ValidateCurrency(currency); err != nil {
		errors.Add("currency", err.Error())
	}

	if err := ValidatePaymentMethod(paymentMethod); err != nil {
		errors.Add("paymentMethod", err.Error())
	}

	return errors
}

// ValidateWalletWithdrawal validates wallet withdrawal request
func ValidateWalletWithdrawal(amount float64, currency, withdrawalMethod string, accountDetails map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	if err := ValidateAmount(amount); err != nil {
		errors.Add("amount", err.Error())
	}

	if err := ValidateCurrency(currency); err != nil {
		errors.Add("currency", err.Error())
	}

	if err := ValidatePaymentMethod(withdrawalMethod); err != nil {
		errors.Add("withdrawalMethod", err.Error())
	}

	if len(accountDetails) == 0 {
		errors.Add("accountDetails", "account details are required")
	}

	return errors
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(input string) string {
	// Remove leading and trailing whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	return input
}

// IsValidID checks if an ID string is valid (non-empty and reasonable length)
func IsValidID(id string) bool {
	if id == "" {
		return false
	}

	if len(id) < 10 || len(id) > 100 {
		return false
	}

	// Check if it contains only alphanumeric characters and allowed special chars
	idRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return idRegex.MatchString(id)
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 30 {
		return fmt.Errorf("username is too long (max 30 characters)")
	}

	// Username can contain letters, numbers, underscores, and hyphens
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}
