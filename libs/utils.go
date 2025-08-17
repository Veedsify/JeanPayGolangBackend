package libs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RoundCurrency rounds a float64 to 2 decimal places for currency
func RoundCurrency(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// CalculateFee calculates fee based on amount and fee percentage
func CalculateFee(amount, feePercentage float64) float64 {
	return RoundCurrency((amount * feePercentage) / 100)
}

// GenerateUUID generates a new UUID without hyphens
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

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPhoneNumber validates phone number format
func IsValidPhoneNumber(phone string) bool {
	// Remove any spaces or dashes
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(cleanPhone)
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(input string) string {
	// Remove leading and trailing whitespace
	input = strings.TrimSpace(input)
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}

// FormatCurrency formats a float64 as currency string
func FormatCurrency(amount float64, currency string) string {
	switch currency {
	case "NGN":
		return fmt.Sprintf("₦%.2f", amount)
	case "GHS":
		return fmt.Sprintf("GH₵%.2f", amount)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
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
	return fmt.Sprintf("%d-%02d-%02d : %02d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
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

// ContainsString checks if a slice contains a specific string
func ContainsString(slice []string, item string) bool {
	return SliceContains(slice, item)
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

// ConvertStringToFloat converts string to float64 with validation
func ConvertStringToFloat(str string) (float64, error) {
	if str == "" {
		return 0, fmt.Errorf("empty string cannot be converted to float")
	}

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

	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid integer format: %s", str)
	}

	return value, nil
}

// ConvertStringToUint32 converts string to uint32 with validation
func ConvertStringToUint32(str string) (uint32, error) {
	if str == "" {
		return 0, fmt.Errorf("empty string cannot be converted to uint32")
	}

	value, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid uint32 format: %s", str)
	}

	return uint32(value), nil
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// IsEmptyString checks if a string is empty or contains only whitespace
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// GetEnvOrDefault gets environment variable with fallback
func GetEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvIntOrDefault gets environment variable as integer with fallback
func GetEnvIntOrDefault(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnvBoolOrDefault gets environment variable as boolean with fallback
func GetEnvBoolOrDefault(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
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

// SliceContains checks if a slice contains an element
func SliceContains[T comparable](slice []T, element T) bool {
	return SliceContains(slice, element)
}

// SliceMap applies a function to each element of a slice
func SliceMap[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// SliceFilter filters a slice based on a predicate function
func SliceFilter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

func SecureRandomNumber(length int) (int64, error) {
	if length <= 0 || length > 18 {
		return 0, fmt.Errorf("Length must be between 1 and 18")
	}
	min := int64Pow(10, length-1)
	max := int64Pow(10, length) - 1

	nBig, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, err
	}
	return nBig.Int64() + min, nil
}

func int64Pow(a, b int) int64 {
	result := int64(1)
	for range b {
		result *= int64(a)
	}
	return result
}

func GenerateUniqueWalletId() (uint64, uint64) {
	var NGN, GHS int64
	var err error
	if NGN, err = SecureRandomNumber(12); err != nil {
		return 0, 0
	}
	if GHS, err = SecureRandomNumber(12); err != nil {
		return 0, 0
	}
	return uint64(NGN), uint64(GHS)
}

func GetDefaultCurrency(country string) string {
	switch country {
	case "NG":
		return "NGN"
	case "GH":
		return "GHS"
	default:
		return "NGN"
	}
}
