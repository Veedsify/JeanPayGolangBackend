package controllers

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/gin-gonic/gin"
)

// GetUserSettingsEndpoint retrieves user settings
func GetUserSettingsEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	settings, err := services.GetUserSettings(userID.(uint32))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User settings retrieved successfully",
		"data":    settings,
	})
}

// Update User Profile Image
func UpdateUserProfileImageEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	userId := claims.(*libs.JWTClaims).ID

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid image file",
		})
		return
	}

	imageURL, err := services.UpdateUserProfileImage(userId, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	data := map[string]interface{}{
		"imageUrl": imageURL,
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Profile image updated successfully",
		"data":    data,
	})
}

// UpdateUserSettingsEndpoint updates user profile information
func UpdateUserSettingsEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	var req services.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	settings, err := services.UpdateUserSettings(userID.(uint32), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Settings updated successfully",
		"data":    settings,
	})
}

// ChangePasswordEndpoint changes user password
func ChangePasswordEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	err := services.ChangeUserPassword(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Password changed successfully",
	})
}

// GetUserPreferencesEndpoint retrieves user preferences
func GetUserPreferencesEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	preferences, err := services.GetUserPreferences(userID.(uint32))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "User preferences retrieved successfully",
		"data":    preferences,
	})
}

// UpdateUserPreferencesEndpoint updates user preferences
func UpdateUserPreferencesEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	var req services.UserPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	preferences, err := services.UpdateUserPreferences(userID.(uint32), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Preferences updated successfully",
		"data":    preferences,
	})
}

// EnableTwoFactorAuthEndpoint enables two-factor authentication
func EnableTwoFactorAuthEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID

	err := services.EnableTwoFactorAuth(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Two-factor authentication enabled successfully",
	})
}

// DisableTwoFactorAuthEndpoint disables two-factor authentication
func DisableTwoFactorAuthEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID

	err := services.DisableTwoFactorAuth(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Two-factor authentication disabled successfully",
	})
}

// DeactivateAccountEndpoint deactivates user account
func DeactivateAccountEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	err := services.DeactivateAccount(userID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Account deactivated successfully",
	})
}

// GetSecuritySettingsEndpoint retrieves security settings
func GetSecuritySettingsEndpoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	// Get user from database to access full user model
	var user models.User
	if err := database.DB.Where("user_id = ?", userID.(uint32)).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to retrieve user information",
		})
		return
	}

	securitySettings := gin.H{
		"twoFactorEnabled":   user.IsTwoFactorEnabled,
		"isVerified":         user.IsVerified,
		"isBlocked":          user.IsBlocked,
		"lastPasswordChange": user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Security settings retrieved successfully",
		"data":    securitySettings,
	})
}

// UpdatePreferencesEndpoint updates user preferences
func UpdatePreferencesEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.UserPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	preferences, err := services.UpdateUserPreferences(claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Preferences updated successfully",
		"data":    preferences,
	})
}

// UpdateProfileEndpoint updates user profile information
func UpdateProfileEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	profile, err := services.UpdateUserProfile(claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Profile updated successfully",
		"data":    profile,
	})
}

// UpdateSecuritySettingsEndpoint updates security settings
func UpdateSecuritySettingsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.UpdateSecuritySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	settings, err := services.UpdateSecuritySettings(claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Security settings updated successfully",
		"data":    settings,
	})
}

// UpdateNotificationSettingsEndpoint updates notification settings
func UpdateNotificationSettingsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	settings, err := services.UpdateNotificationSettings(claims.UserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Notification settings updated successfully",
		"data":    settings,
	})
}

// GenerateTwoFactorQREndpoint generates QR code for two-factor authentication
func GenerateTwoFactorQREndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	qrData, err := services.GenerateTwoFactorQR(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Two-factor QR code generated successfully",
		"data":    qrData,
	})
}

// EnableTwoFactorEndpoint enables two-factor authentication
func EnableTwoFactorEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.EnableTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := services.EnableTwoFactorAuthentication(claims.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Two-factor authentication enabled successfully",
	})
}

// DisableTwoFactorEndpoint disables two-factor authentication
func DisableTwoFactorEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	var req services.DisableTwoFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	err := services.DisableTwoFactorAuthentication(claims.ID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Two-factor authentication disabled successfully",
	})
}
