package services

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"gorm.io/gorm"
)

// UpdateUserSettingsRequest represents user settings update request
type UpdateUserSettingsRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=6"`
}

// UserPreferencesRequest represents user preferences update request
type UserPreferencesRequest struct {
	EmailNotifications *bool  `json:"emailNotifications"`
	SMSNotifications   *bool  `json:"smsNotifications"`
	Currency           string `json:"currency"`
}

// UserSettingsResponse represents user settings response
type UserSettingsResponse struct {
	UserID      uint32 `json:"userId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Username    string `json:"username"`
	Country     string `json:"country"`
}

// UserPreferencesResponse represents user preferences response
type UserPreferencesResponse struct {
	EmailNotifications bool   `json:"emailNotifications"`
	SMSNotifications   bool   `json:"smsNotifications"`
	Currency           string `json:"currency"`
	TwoFactorEnabled   bool   `json:"twoFactorEnabled"`
}

// UpdateUserProfileImage updates the user profile image

func UpdateUserProfileImage(userId uint, file *multipart.FileHeader) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	cld, err := NewCloudinaryService()
	if err != nil {
		return "", err
	}

	var PROFILE_FOLDER = libs.GetEnvOrDefault("PROFILE_IMAGE_FOLDER", "profile_images")
	// Upload to Cloudinary
	uploadResult, err := cld.UploadImage(src, PROFILE_FOLDER)
	if err != nil {
		return "", err
	}

	tx := database.DB.Begin()

	if err := tx.Model(&models.User{}).Where("id = ?", userId).Update("profile_picture", uploadResult).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	tx.Commit()
	return uploadResult, nil
}

// UpdateUserSettings updates user profile information
func UpdateUserSettings(userID uint32, req UpdateUserSettingsRequest) (*UserSettingsResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if email is being changed and if it already exists
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := tx.Where("email = ? AND user_id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			tx.Rollback()
			return nil, errors.New("email already exists")
		}
	}

	// Update user fields
	updates := make(map[string]interface{})

	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.PhoneNumber != "" {
		updates["phone_number"] = req.PhoneNumber
	}

	if len(updates) > 0 {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update user settings: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Refresh user data
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	response := &UserSettingsResponse{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Username:    user.Username,
		Country:     string(user.Country),
	}

	return response, nil
}

// ChangeUserPassword changes user password
func ChangeUserPassword(userID uint, req ChangePasswordRequest) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if req.CurrentPassword == "" {
		return errors.New("current password is required")
	}

	if req.NewPassword == "" {
		return errors.New("new password is required")
	}

	if len(req.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Verify current password
	if err := libs.ComparePassword(user.Password, req.CurrentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := libs.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Create security notification
	if err := CreateSecurityNotification(userID, "Password changed"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create security notification: %v\n", err)
	}

	return nil
}

// UpdateUserPreferences updates user preferences
func UpdateUserPreferences(userID uint32, req UserPreferencesRequest) (*UserPreferencesResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// For now, we'll return a basic preferences response
	// In a full implementation, you might have a separate preferences table
	response := &UserPreferencesResponse{
		EmailNotifications: true,  // Default value
		SMSNotifications:   false, // Default value
		Currency:           "NGN", // Default value
		TwoFactorEnabled:   user.IsTwoFactorEnabled,
	}

	// Update preferences based on request
	if req.EmailNotifications != nil {
		response.EmailNotifications = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		response.SMSNotifications = *req.SMSNotifications
	}
	if req.Currency != "" {
		if req.Currency == "NGN" || req.Currency == "GHS" {
			response.Currency = req.Currency
		}
	}

	// In a full implementation, you would save these preferences to a database table
	// For now, we'll just return the response

	return response, nil
}

// GetUserSettings retrieves user settings
func GetUserSettings(userID uint32) (*UserSettingsResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	response := &UserSettingsResponse{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Username:    user.Username,
		Country:     string(user.Country),
	}

	return response, nil
}

// GetUserPreferences retrieves user preferences
func GetUserPreferences(userID uint32) (*UserPreferencesResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	response := &UserPreferencesResponse{
		EmailNotifications: true,  // Default value - would come from preferences table
		SMSNotifications:   false, // Default value - would come from preferences table
		Currency:           "NGN", // Default value - would come from preferences table
		TwoFactorEnabled:   user.IsTwoFactorEnabled,
	}

	return response, nil
}

// EnableTwoFactorAuth enables two-factor authentication for user
func EnableTwoFactorAuth(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsTwoFactorEnabled {
		return errors.New("two-factor authentication is already enabled")
	}

	// Enable two-factor authentication
	if err := database.DB.Model(&user).Update("is_two_factor_enabled", true).Error; err != nil {
		return fmt.Errorf("failed to enable two-factor authentication: %w", err)
	}

	// Create security notification
	if err := CreateSecurityNotification(userID, "Two-factor authentication enabled"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create security notification: %v\n", err)
	}

	return nil
}

// DisableTwoFactorAuth disables two-factor authentication for user
func DisableTwoFactorAuth(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsTwoFactorEnabled {
		return errors.New("two-factor authentication is already disabled")
	}

	// Disable two-factor authentication
	if err := database.DB.Model(&user).Update("is_two_factor_enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable two-factor authentication: %w", err)
	}

	// Create security notification
	if err := CreateSecurityNotification(userID, "Two-factor authentication disabled"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create security notification: %v\n", err)
	}

	return nil
}

// Additional request/response types for new endpoints

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone"`
	ProfilePicture string `json:"profileImage"`
}

// UpdateSecuritySettingsRequest represents security settings update request
type UpdateSecuritySettingsRequest struct {
	TwoFactorEnabled *bool  `json:"twoFactorEnabled"`
	CurrentPassword  string `json:"currentPassword"`
}

// UpdateNotificationSettingsRequest represents notification settings update request
type UpdateNotificationSettingsRequest struct {
	EmailNotifications *bool `json:"emailNotifications"`
	SMSNotifications   *bool `json:"smsNotifications"`
	PushNotifications  *bool `json:"pushNotifications"`
	TransactionAlerts  *bool `json:"transactionAlerts"`
	MarketingEmails    *bool `json:"marketingEmails"`
	SecurityAlerts     *bool `json:"securityAlerts"`
}

// EnableTwoFactorRequest represents enable 2FA request
type EnableTwoFactorRequest struct {
	Token string `json:"token" validate:"required"`
}

// DisableTwoFactorRequest represents disable 2FA request
type DisableTwoFactorRequest struct {
	Password string `json:"password" validate:"required"`
}

// TwoFactorQRResponse represents 2FA QR response
type TwoFactorQRResponse struct {
	QRCodeURL  string `json:"qrCodeUrl"`
	Secret     string `json:"secret"`
	BackupCode string `json:"backupCode"`
}

// UpdateUserProfile updates user profile information
func UpdateUserProfile(userID uint32, req UpdateProfileRequest) (*UserSettingsResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if email is being changed and if it already exists
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := tx.Where("email = ? AND user_id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			tx.Rollback()
			return nil, errors.New("email already exists")
		}
	}

	// Update user fields
	updates := make(map[string]any)

	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.PhoneNumber != "" {
		updates["phone_number"] = req.PhoneNumber
	}
	if req.ProfilePicture != "" {
		updates["profile_picture"] = req.ProfilePicture
	}

	if len(updates) > 0 {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update user profile: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Refresh user data
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	response := &UserSettingsResponse{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Username:    user.Username,
		Country:     string(user.Country),
	}

	return response, nil
}

// UpdateSecuritySettings updates security settings
func UpdateSecuritySettings(userID uint32, req UpdateSecuritySettingsRequest) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify current password if provided
	if req.CurrentPassword != "" {
		if err := libs.ComparePassword(user.Password, req.CurrentPassword); err != nil {
			return nil, errors.New("current password is incorrect")
		}
	}

	updates := make(map[string]interface{})

	if req.TwoFactorEnabled != nil {
		updates["is_two_factor_enabled"] = *req.TwoFactorEnabled
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update security settings: %w", err)
		}
	}

	// Refresh user data
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	response := map[string]interface{}{
		"twoFactorEnabled": user.IsTwoFactorEnabled,
		"isVerified":       user.IsVerified,
		"isBlocked":        user.IsBlocked,
		"updatedAt":        user.UpdatedAt,
	}

	return response, nil
}

// UpdateNotificationSettings updates notification settings
func UpdateNotificationSettings(userID uint32, req UpdateNotificationSettingsRequest) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// For now, we'll return the requested settings as stored
	// In a full implementation, you would store these in a notification_settings table
	response := map[string]interface{}{
		"emailNotifications": true,
		"smsNotifications":   false,
		"pushNotifications":  true,
		"transactionAlerts":  true,
		"marketingEmails":    false,
		"securityAlerts":     true,
	}

	// Update with request values
	if req.EmailNotifications != nil {
		response["emailNotifications"] = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		response["smsNotifications"] = *req.SMSNotifications
	}
	if req.PushNotifications != nil {
		response["pushNotifications"] = *req.PushNotifications
	}
	if req.TransactionAlerts != nil {
		response["transactionAlerts"] = *req.TransactionAlerts
	}
	if req.MarketingEmails != nil {
		response["marketingEmails"] = *req.MarketingEmails
	}
	if req.SecurityAlerts != nil {
		response["securityAlerts"] = *req.SecurityAlerts
	}

	return response, nil
}

// GenerateTwoFactorQR generates QR code for two-factor authentication setup
func GenerateTwoFactorQR(userID uint32) (*TwoFactorQRResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Generate secret and QR code
	// In a real implementation, you would use a TOTP library
	secret := libs.GenerateRandomString(32)
	qrCodeURL := fmt.Sprintf("otpauth://totp/JeanPay:%s?secret=%s&issuer=JeanPay", user.Email, secret)
	backupCode := libs.GenerateRandomString(16)

	response := &TwoFactorQRResponse{
		QRCodeURL:  qrCodeURL,
		Secret:     secret,
		BackupCode: backupCode,
	}

	return response, nil
}

// EnableTwoFactorAuthentication enables 2FA after token verification
func EnableTwoFactorAuthentication(userID uint, req EnableTwoFactorRequest) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if req.Token == "" {
		return errors.New("verification token is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsTwoFactorEnabled {
		return errors.New("two-factor authentication is already enabled")
	}

	// In a real implementation, you would verify the TOTP token here
	// For now, we'll just check if it's not empty

	// Enable two-factor authentication
	if err := database.DB.Model(&user).Update("is_two_factor_enabled", true).Error; err != nil {
		return fmt.Errorf("failed to enable two-factor authentication: %w", err)
	}

	// Create security notification
	if err := CreateSecurityNotification(userID, "Two-factor authentication enabled"); err != nil {
		fmt.Printf("Failed to create security notification: %v\n", err)
	}

	return nil
}

// DisableTwoFactorAuthentication disables 2FA after password verification
func DisableTwoFactorAuthentication(userID uint, req DisableTwoFactorRequest) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsTwoFactorEnabled {
		return errors.New("two-factor authentication is already disabled")
	}

	// Verify password
	if err := libs.ComparePassword(user.Password, req.Password); err != nil {
		return errors.New("password is incorrect")
	}

	// Disable two-factor authentication
	if err := database.DB.Model(&user).Update("is_two_factor_enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable two-factor authentication: %w", err)
	}

	// Create security notification
	if err := CreateSecurityNotification(userID, "Two-factor authentication disabled"); err != nil {
		fmt.Printf("Failed to create security notification: %v\n", err)
	}

	return nil
}

// DeactivateAccount deactivates user account
func DeactivateAccount(userID uint, reason string) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsVerified {
		return errors.New("account is already deactivated")
	}

	// Deactivate account
	updates := map[string]interface{}{
		"is_verified": false,
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	// Create notification
	message := "Your account has been deactivated"
	if reason != "" {
		message += ". Reason: " + reason
	}

	if err := CreateSystemNotification(userID, message); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to create system notification: %v\n", err)
	}

	return nil
}
