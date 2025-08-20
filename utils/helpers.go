package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID without hyphenswetin
func GenerateUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenerateTransactionReference generates a unique transaction reference
func GenerateTransactionReference(prefix string) string {
	timestamp := time.Now().Unix()
	randomPart := GenerateRandomString(8)
	return fmt.Sprintf("%s_%d_%s", prefix, timestamp, randomPart)
}

// GenerateRandomString generates a random alphanumeric string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func GetConvertdirection(currency string) models.TransactionDirection {
	switch currency {
	case "NGN":
		return models.NGNToGHS
	case "GHS":
		return models.GHSToNGN
	default:
		return ""
	}
}

// GenerateRandomNumericString generates a random numeric string of specified length
func GenerateRandomNumericString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

// GenerateOTP generates a One-Time Password of specified length
func GenerateOTP(length int) string {
	if length <= 0 {
		length = 6
	}
	return GenerateRandomNumericString(length)
}

// HashString creates a SHA256 hash of the input string
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// RoundFloat rounds a float64 to specified decimal places
func RoundFloat(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// RoundCurrency rounds a float64 to 2 decimal places for currency
func RoundCurrency(amount float64) float64 {
	return RoundFloat(amount, 2)
}

// CalculatePercentage calculates percentage of a number
func CalculatePercentage(value, percentage float64) float64 {
	return RoundCurrency((value * percentage) / 100)
}

// CalculateFee calculates fee based on amount and fee percentage
func CalculateFee(amount, feePercentage float64) float64 {
	return CalculatePercentage(amount, feePercentage)
}

// ConvertStringToFloat converts string to float64 with validation
func ConvertStringToFloat(str string) (float64, error) {
	if str == "" {
		return 0, fmt.Errorf("empty string cannot be converted to float")
	}

	str = strings.ReplaceAll(strings.TrimSpace(str), ",", "")

	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", str)
	}

	return value, nil
}

// ConvertStringToInt converts string to int with validation
func ConvertStringToInt(str string) (int, error) {
	if str == "" {
		return 0, fmt.Errorf("empty string cannot be converted to int")
	}
	str = strings.ReplaceAll(strings.TrimSpace(str), ",", "")

	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid integer format: %s", str)
	}

	return value, nil
}

// IsEmptyString checks if a string is empty or contains only whitespace
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// ContainsString checks if a slice contains a specific string
func ContainsString(slice []string, item string) bool {
	return strings.Contains(strings.Join(slice, "\n"), item)
}

// RemoveString removes all occurrences of a string from a slice
func RemoveString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings removes duplicate strings from a slice
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// MaskEmail masks an email address for privacy
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return email
	}

	maskedUsername := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	return maskedUsername + "@" + domain
}

// MaskPhoneNumber masks a phone number for privacy
func MaskPhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return phone
	}

	return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
}

// FormatCurrency formats a float64 as currency string
func FormatCurrency(amount float64, currency string) string {
	formattedAmount := humanize.CommafWithDigits(amount, 2)
	switch currency {
	case "NGN":
		return fmt.Sprintf("₦ %s", formattedAmount)
	case "GHS":
		return fmt.Sprintf("GH₵ %s", formattedAmount)
	default:
		return fmt.Sprintf("%s %s", formattedAmount, currency)
	}
}

// ParseCurrencyAmount extracts numeric amount from currency string
func ParseCurrencyAmount(currencyStr string) (float64, error) {
	// Remove currency symbols and spaces
	re := regexp.MustCompile(`[^\d.]`)
	cleanStr := re.ReplaceAllString(currencyStr, "")

	if cleanStr == "" {
		return 0, fmt.Errorf("no numeric value found in currency string")
	}

	return ConvertStringToFloat(cleanStr)
}

// TimeAgo returns a human-readable time difference
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < time.Hour*24 {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < time.Hour*24*7 {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < time.Hour*24*30 {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < time.Hour*24*365 {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// FormatTime formats time in a standard format
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate formats date in a standard format
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseDate parses date string in various formats
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"02/01/2006",
		"02-01-2006",
		"January 2, 2006",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// IsWeekend checks if a given time is on a weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// GetBusinessDays calculates number of business days between two dates
func GetBusinessDays(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}

	days := 0
	for d := start; d.Before(end) || d.Equal(end); d = d.AddDate(0, 0, 1) {
		if !IsWeekend(d) {
			days++
		}
	}

	return days
}

// StructToMap converts a struct to a map[string]any
func StructToMap(obj any) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		jsonTag := fieldType.Tag.Get("json")
		fieldName := fieldType.Name

		if jsonTag != "" && jsonTag != "-" {
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				fieldName = jsonTag[:idx]
			} else {
				fieldName = jsonTag
			}
		}

		result[fieldName] = field.Interface()
	}

	return result
}

// DeepCopy creates a deep copy of a map
func DeepCopyMap(original map[string]any) map[string]any {
	copy := make(map[string]any)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

// MergeStringMaps merges two string maps, with the second map taking precedence
func MergeStringMaps(map1, map2 map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range map1 {
		result[k] = v
	}

	for k, v := range map2 {
		result[k] = v
	}

	return result
}

// GetMapKeys returns all keys from a map[string]any
func GetMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// SafeStringAccess safely accesses a string value from a map
func SafeStringAccess(m map[string]any, key string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// SafeFloatAccess safely accesses a float64 value from a map
func SafeFloatAccess(m map[string]any, key string) float64 {
	if value, exists := m[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := ConvertStringToFloat(v); err == nil {
				return f
			}
		}
	}
	return 0
}

// SafeIntAccess safely accesses an int value from a map
func SafeIntAccess(m map[string]any, key string) int {
	if value, exists := m[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := ConvertStringToInt(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// SafeBoolAccess safely accesses a bool value from a map
func SafeBoolAccess(m map[string]any, key string) bool {
	if value, exists := m[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// GetStringOrDefault gets a string value or returns default
func GetStringOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntOrDefault gets an int value or returns default
func GetIntOrDefault(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// GetFloatOrDefault gets a float64 value or returns default
func GetFloatOrDefault(value, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

type CommonError struct {
	code        string
	title       string
	description string
	action      string
}

var commonErrors = []CommonError{
	{
		code:        "INSUFFICIENT_FUNDS",
		title:       "Insufficient Funds",
		description: "Your account balance is not sufficient for this transfer.",
		action:      "Add funds to your account and try again.",
	},
	{
		code:        "INVALID_AMOUNTS",
		title:       "Invalid Amounts",
		description: "The amounts provided for the transfer are invalid.",
		action:      "Please check the amounts and try again.",
	},
	{
		code:        "INVALID_ACCOUNT",
		title:       "Invalid Account Details",
		description: "The recipient account details could not be verified.",
		action:      "Please check the account details and try again.",
	},
	{
		code:        "NETWORK_ERROR",
		title:       "Network Connection Error",
		description: "There was a problem connecting to our payment network.",
		action:      "Check your internet connection and try again.",
	},
	{
		code:        "RATE_LIMIT",
		title:       "Too Many Attempts",
		description: "You have exceeded the maximum number of transfer attempts.",
		action:      "Please wait a few minutes before trying again.",
	},
	{
		code:        "MAINTENANCE",
		title:       "Service Temporarily Unavailable",
		description: "Our payment service is currently under maintenance.",
		action:      "Please try again in a few minutes.",
	},
	{
		code:        "INTERNAL_SERVER_ERROR",
		title:       "Internal Server Error",
		description: "An unexpected error occurred on our server.",
		action:      "Please try again later or contact support if the issue persists.",
	},
	{
		code:        "NO_PAYMENT_RECEIVED",
		title:       "No Payment Received",
		description: "No payment was received for this transaction.",
		action:      "Please check your payment method and try again.",
	},
	{
		code:        "TRANSACTION_NOT_FOUND",
		title:       "Transaction Not Found",
		description: "The transaction you are looking for does not exist.",
		action:      "Please check the transaction ID and try again.",
	},
}

func GetErrorFromCode(code string) string {
	for _, err := range commonErrors {
		if strings.EqualFold(err.code, code) {
			jsonStr := fmt.Sprintf(`{"code":"%s","title":"%s","description":"%s","action":"%s"}`, err.code, err.title, err.description, err.action)
			return jsonStr
		}
	}
	return `{}`
}
