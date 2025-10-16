package utils

import (
	"errors"
	"fmt"
	"strings"
)

// CurrencyPair represents a supported currency conversion pair
type CurrencyPair struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// String returns string representation of currency pair
func (cp CurrencyPair) String() string {
	return fmt.Sprintf("%s-%s", cp.From, cp.To)
}

// SUPPORTED_PAIRS defines the 7 supported currency conversion pairs
var SUPPORTED_PAIRS = []CurrencyPair{
	{From: "NGN", To: "GHS"},
	{From: "GHS", To: "NGN"},
	{From: "USD", To: "GHS"},
	{From: "NGN", To: "XOF"},
	{From: "GHS", To: "XOF"},
	{From: "XOF", To: "GHS"},
	{From: "XOF", To: "NGN"},
}

// XOF_SUPPORTED_COUNTRIES defines the 8 West African countries that support XOF currency
var XOF_SUPPORTED_COUNTRIES = []string{
	"BJ", // Benin
	"BF", // Burkina Faso
	"CI", // Côte d'Ivoire
	"GW", // Guinea-Bissau
	"ML", // Mali
	"NE", // Niger
	"SN", // Senegal
	"TG", // Togo
}

// ValidateCurrencyPair validates if a currency pair is supported
func ValidateCurrencyPair(fromCurrency, toCurrency string) error {
	if fromCurrency == "" {
		return errors.New("from currency is required")
	}

	if toCurrency == "" {
		return errors.New("to currency is required")
	}

	if fromCurrency == toCurrency {
		return errors.New("cannot convert to the same currency")
	}

	// Check if the pair is in supported pairs
	for _, pair := range SUPPORTED_PAIRS {
		if pair.From == fromCurrency && pair.To == toCurrency {
			return nil
		}
	}

	// Generate list of supported pairs for error message
	var supportedPairStrings []string
	for _, pair := range SUPPORTED_PAIRS {
		supportedPairStrings = append(supportedPairStrings, pair.String())
	}

	return fmt.Errorf("unsupported currency pair %s-%s. Supported pairs: %s",
		fromCurrency, toCurrency, strings.Join(supportedPairStrings, ", "))
}

// IsXOFCountrySupported validates if a country supports XOF currency
func IsXOFCountrySupported(countryCode string) bool {
	if countryCode == "" {
		return false
	}

	// Convert to uppercase for comparison
	countryCode = strings.ToUpper(countryCode)

	for _, supportedCountry := range XOF_SUPPORTED_COUNTRIES {
		if countryCode == supportedCountry {
			return true
		}
	}

	return false
}

// ValidateXOFCountry validates XOF currency usage for specific countries
func ValidateXOFCountry(currency, countryCode string) error {
	if currency != "XOF" {
		return nil // No validation needed for non-XOF currencies
	}

	if countryCode == "" {
		return errors.New("country code is required when using XOF currency")
	}

	if !IsXOFCountrySupported(countryCode) {
		supportedCountries := strings.Join(XOF_SUPPORTED_COUNTRIES, ", ")
		return fmt.Errorf("XOF currency is only supported in these countries: %s", supportedCountries)
	}

	return nil
}

// GetSupportedCurrencyPairs returns all supported currency pairs
func GetSupportedCurrencyPairs() []CurrencyPair {
	return SUPPORTED_PAIRS
}

// GetXOFSupportedCountries returns all countries that support XOF
func GetXOFSupportedCountries() []string {
	return XOF_SUPPORTED_COUNTRIES
}

// IsCurrencySupported checks if a currency is supported in any pair
func IsCurrencySupported(currency string) bool {
	for _, pair := range SUPPORTED_PAIRS {
		if pair.From == currency || pair.To == currency {
			return true
		}
	}
	return false
}

// GetSupportedCurrenciesFromPairs returns unique list of all supported currencies
func GetSupportedCurrenciesFromPairs() []string {
	currencySet := make(map[string]bool)

	for _, pair := range SUPPORTED_PAIRS {
		currencySet[pair.From] = true
		currencySet[pair.To] = true
	}

	var currencies []string
	for currency := range currencySet {
		currencies = append(currencies, currency)
	}

	return currencies
}
