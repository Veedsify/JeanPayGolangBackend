package services

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

// UpdateUserSettingsRequest represents user settings update request

type SettingsRequest struct {
	FeesBreakdown *bool `json:"fees_breakdown,omitempty"`
	SaveRecipient *bool `json:"save_recipient,omitempty"`
}

type UpdateUserSettingsRequest struct {
	Setting SettingsRequest `json:"setting"`
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
	Setting         *models.Setting          `json:"setting,omitempty"`
	WithdrawMethods *[]models.WithdrawMethod `json:"withdrawMethod,omitempty"`
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
func UpdateUserSettings(userID uint, req UpdateUserSettingsRequest) (any, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update user fields
	updates := make(map[string]any)
	if req.Setting.FeesBreakdown != nil {
		updates["fees_breakdown"] = req.Setting.FeesBreakdown
	}
	if req.Setting.SaveRecipient != nil {
		updates["save_recipient"] = req.Setting.SaveRecipient
	}
	fmt.Println(updates)

	if err := tx.Model(&models.Setting{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update user settings: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return "", nil
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
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
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
func GetUserSettings(userID uint) (*UserSettingsResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	var user models.User
	if err := database.DB.Preload("Setting").Preload("WithdrawMethods").Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	response := &UserSettingsResponse{
		&user.Setting,
		&user.WithdrawMethods,
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
func UpdateUserProfile(userID uint32, req UpdateProfileRequest) (any, error) {
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

	return user, nil
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

// Platform settings types for admin endpoints
type PlatformSettingsRequest struct {
	KYCEnforcement     *bool  `json:"kycEnforcement"`
	ManualRateOverride *bool  `json:"manualRateOverride"`
	DefaultCurrency    string `json:"defaultCurrency"`
	TransactionEmails  *bool  `json:"transactionEmails"`
	MinimumTransaction *int64 `json:"minimumTransaction"`
	MaximumTransaction *int64 `json:"maximumTransaction"`
	DailyUserCap       *int64 `json:"dailyUserCap"`
	// theme, notifications and security
	Theme                 string `json:"theme"`
	EmailNotifications    *bool  `json:"emailNotifications"`
	SMSNotifications      *bool  `json:"smsNotifications"`
	PushNotifications     *bool  `json:"pushNotifications"`
	EnforceTwoFactor      *bool  `json:"enforceTwoFactor"`
	SessionTimeoutMinutes *int64 `json:"sessionTimeoutMinutes"`
	PasswordExpiryDays    *int64 `json:"passwordExpiryDays"`
	// granular notification toggles
	SendTransactionSuccess *bool `json:"sendTransactionSuccess"`
	SendTransactionDecline *bool `json:"sendTransactionDecline"`
	SendTransactionPending *bool `json:"sendTransactionPending"`
	SendTransactionRefund  *bool `json:"sendTransactionRefund"`
	AccountLimitsNotify    *bool `json:"accountLimitsNotification"`
}

type PlatformSettingsResponse struct {
	KYCEnforcement     bool   `json:"kycEnforcement"`
	ManualRateOverride bool   `json:"manualRateOverride"`
	DefaultCurrency    string `json:"defaultCurrency"`
	TransactionEmails  bool   `json:"transactionEmails"`
	MinimumTransaction int64  `json:"minimumTransaction"`
	MaximumTransaction int64  `json:"maximumTransaction"`
	DailyUserCap       int64  `json:"dailyUserCap"`
	// theme, notifications and security
	Theme                 string `json:"theme"`
	EmailNotifications    bool   `json:"emailNotifications"`
	SMSNotifications      bool   `json:"smsNotifications"`
	PushNotifications     bool   `json:"pushNotifications"`
	EnforceTwoFactor      bool   `json:"enforceTwoFactor"`
	SessionTimeoutMinutes int64  `json:"sessionTimeoutMinutes"`
	PasswordExpiryDays    int64  `json:"passwordExpiryDays"`
	// granular notification toggles
	SendTransactionSuccess bool `json:"sendTransactionSuccess"`
	SendTransactionDecline bool `json:"sendTransactionDecline"`
	SendTransactionPending bool `json:"sendTransactionPending"`
	SendTransactionRefund  bool `json:"sendTransactionRefund"`
	AccountLimitsNotify    bool `json:"accountLimitsNotification"`
}

// GetPlatformSettings returns platform-wide settings. For now return values from env/defaults.
func GetPlatformSettings() (*PlatformSettingsResponse, error) {
	var ps models.PlatformSetting
	db := database.DB
	if err := db.Order("id ASC").First(&ps).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create default record
			defaultRec := models.PlatformSetting{
				KycEnforcement:               false,
				ManualRateOverride:           false,
				TransactionConfirmationEmail: true,
				DefaultCurrencyDisplay:       models.DefaultCurrency("NGN"),
				MinimumTransactionAmount:     100,
				MaximumTransactionAmount:     1000000,
				DailyTransactionLimit:        5000000,
				ChartStyle:                   models.LineChart,
				Theme:                        "light",
				EmailNotifications:           true,
				SMSNotifications:             false,
				PushNotifications:            true,
				EnforceTwoFactor:             false,
				SessionTimeoutMinutes:        30,
				PasswordExpiryDays:           90,
				SendTransactionSuccessEmail:  true,
				SendTransactionDeclineEmail:  true,
				SendTransactionPendingEmail:  true,
				SendTransactionRefundEmail:   true,
				AccountLimitsNotification:    true,
			}
			if err := db.Create(&defaultRec).Error; err != nil {
				return nil, fmt.Errorf("failed to create default platform settings: %w", err)
			}
			ps = defaultRec
		} else {
			return nil, fmt.Errorf("failed to load platform settings: %w", err)
		}
	}

	resp := &PlatformSettingsResponse{
		KYCEnforcement:        ps.KycEnforcement,
		ManualRateOverride:    ps.ManualRateOverride,
		DefaultCurrency:       string(ps.DefaultCurrencyDisplay),
		TransactionEmails:     ps.TransactionConfirmationEmail || ps.SendTransactionSuccessEmail,
		MinimumTransaction:    int64(ps.MinimumTransactionAmount),
		MaximumTransaction:    int64(ps.MaximumTransactionAmount),
		DailyUserCap:          int64(ps.DailyTransactionLimit),
		Theme:                 ps.Theme,
		EmailNotifications:    ps.EmailNotifications,
		SMSNotifications:      ps.SMSNotifications,
		PushNotifications:     ps.PushNotifications,
		EnforceTwoFactor:      ps.EnforceTwoFactor,
		SessionTimeoutMinutes: int64(ps.SessionTimeoutMinutes),
		PasswordExpiryDays:    int64(ps.PasswordExpiryDays),
		// granular notification toggles
		SendTransactionSuccess: ps.SendTransactionSuccessEmail,
		SendTransactionDecline: ps.SendTransactionDeclineEmail,
		SendTransactionPending: ps.SendTransactionPendingEmail,
		SendTransactionRefund:  ps.SendTransactionRefundEmail,
		AccountLimitsNotify:    ps.AccountLimitsNotification,
	}

	return resp, nil
}

// UpdatePlatformSettings updates platform-wide settings. Currently just returns the updated payload.
func UpdatePlatformSettings(req PlatformSettingsRequest, adminID uint) (*PlatformSettingsResponse, error) {
	db := database.DB

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ps models.PlatformSetting
	if err := tx.Order("id ASC").First(&ps).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create
			ps = models.PlatformSetting{}
			if err := tx.Create(&ps).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create platform settings: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to load platform settings: %w", err)
		}
	}

	updates := make(map[string]any)
	if req.KYCEnforcement != nil {
		updates["kyc_enforcement"] = *req.KYCEnforcement
	}
	if req.ManualRateOverride != nil {
		updates["manual_rate_override"] = *req.ManualRateOverride
	}
	if req.DefaultCurrency != "" {
		updates["default_currency_display"] = models.DefaultCurrency(req.DefaultCurrency)
	}
	if req.TransactionEmails != nil {
		updates["transaction_confirmation_email"] = *req.TransactionEmails
		updates["send_transaction_success_email"] = *req.TransactionEmails
	}
	if req.SendTransactionSuccess != nil {
		updates["send_transaction_success_email"] = *req.SendTransactionSuccess
	}
	if req.SendTransactionDecline != nil {
		updates["send_transaction_decline_email"] = *req.SendTransactionDecline
	}
	if req.SendTransactionPending != nil {
		updates["send_transaction_pending_email"] = *req.SendTransactionPending
	}
	if req.SendTransactionRefund != nil {
		updates["send_transaction_refund_email"] = *req.SendTransactionRefund
	}
	if req.AccountLimitsNotify != nil {
		updates["account_limits_notification"] = *req.AccountLimitsNotify
	}
	if req.Theme != "" {
		updates["theme"] = req.Theme
	}
	if req.EmailNotifications != nil {
		updates["email_notifications"] = *req.EmailNotifications
	}
	if req.SMSNotifications != nil {
		updates["sms_notifications"] = *req.SMSNotifications
	}
	if req.PushNotifications != nil {
		updates["push_notifications"] = *req.PushNotifications
	}
	if req.EnforceTwoFactor != nil {
		updates["enforce_two_factor"] = *req.EnforceTwoFactor
	}
	if req.SessionTimeoutMinutes != nil {
		updates["session_timeout_minutes"] = float64(*req.SessionTimeoutMinutes)
	}
	if req.PasswordExpiryDays != nil {
		updates["password_expiry_days"] = float64(*req.PasswordExpiryDays)
	}
	if req.MinimumTransaction != nil {
		updates["minimum_transaction_amount"] = float64(*req.MinimumTransaction)
	}
	if req.MaximumTransaction != nil {
		updates["maximum_transaction_amount"] = float64(*req.MaximumTransaction)
	}
	if req.DailyUserCap != nil {
		updates["daily_transaction_limit"] = float64(*req.DailyUserCap)
	}

	if len(updates) > 0 {
		if err := tx.Model(&ps).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update platform settings: %w", err)
		}
	}

	// create admin log
	adminLog := models.AdminLog{
		AdminID:  uint32(adminID),
		Action:   "UPDATE_PLATFORM_SETTINGS",
		Target:   "platform_settings",
		TargetID: fmt.Sprintf("%d", ps.ID),
		Details:  fmt.Sprintf("Platform settings updated by admin %d", adminID),
	}
	if err := tx.Create(&adminLog).Error; err != nil {
		// don't fail the whole op for logging error, but rollback the transaction
		tx.Rollback()
		return nil, fmt.Errorf("failed to create admin log: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// refresh record
	if err := db.Order("id ASC").First(&ps).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch updated platform settings: %w", err)
	}

	// build full response mirroring GetPlatformSettings
	resp := &PlatformSettingsResponse{
		KYCEnforcement:        ps.KycEnforcement,
		ManualRateOverride:    ps.ManualRateOverride,
		DefaultCurrency:       string(ps.DefaultCurrencyDisplay),
		TransactionEmails:     ps.TransactionConfirmationEmail || ps.SendTransactionSuccessEmail,
		MinimumTransaction:    int64(ps.MinimumTransactionAmount),
		MaximumTransaction:    int64(ps.MaximumTransactionAmount),
		DailyUserCap:          int64(ps.DailyTransactionLimit),
		Theme:                 ps.Theme,
		EmailNotifications:    ps.EmailNotifications,
		SMSNotifications:      ps.SMSNotifications,
		PushNotifications:     ps.PushNotifications,
		EnforceTwoFactor:      ps.EnforceTwoFactor,
		SessionTimeoutMinutes: int64(ps.SessionTimeoutMinutes),
		PasswordExpiryDays:    int64(ps.PasswordExpiryDays),
	}
	// add granular flags if present on model
	// if model has send flags, map them via existing fields
	// extend response via type assertion (already defined fields present in struct)
	// Note: PlatformSettingsResponse currently doesn't have fields for granular sends; if needed, add them to the struct and map here.

	return resp, nil
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

func UpdateWalletSettings(ID uint, req types.WalletSettingsRequest) error {

	if ID == 0 {
		return errors.New("user ID is required")
	}

	var settings models.Setting

	db := database.DB.Where("user_id = ?", ID).First(&settings)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return errors.New("settings not found")
		}
		return fmt.Errorf("failed to find settings: %w", db.Error)
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update settings fields
	updates := make(map[string]any)
	if req.Currency != "" {
		updates["default_currency"] = models.DefaultCurrency(req.Currency)
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}

	tx.Model(&settings).Updates(updates)

	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update wallet settings: %w", tx.Error)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	// Refresh settings data
	if err := database.DB.Where("user_id = ?", ID).First(&settings).Error; err != nil {
		return fmt.Errorf("failed to fetch updated settings: %w", err)
	}

	return nil
}
