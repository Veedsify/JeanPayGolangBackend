package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func SettingsRoutes(router *gin.RouterGroup) {
	settings := router.Group(constants.SettingsBase)
	{
		settings.GET("/", controllers.GetUserSettingsEndpoint)
		settings.PUT(constants.SettingsUpdate, controllers.UpdateUserSettingsEndpoint)
		settings.POST(constants.SettingsProfilePicture, controllers.UpdateUserProfileImageEndpoint)
		settings.PUT(constants.SettingsChangePassword, controllers.ChangePasswordEndpoint)
		settings.PUT(constants.SettingsPreferences, controllers.UpdatePreferencesEndpoint)
		settings.PUT(constants.SettingsProfile, controllers.UpdateProfileEndpoint)
		settings.PUT(constants.SettingsSecurity, controllers.UpdateSecuritySettingsEndpoint)
		settings.PUT(constants.SettingsNotifications, controllers.UpdateNotificationSettingsEndpoint)
		settings.GET(constants.SettingsTwoFactor, controllers.GenerateTwoFactorQREndpoint)
		settings.POST(constants.SettingsTwoFactorEnable, controllers.EnableTwoFactorEndpoint)
		settings.POST(constants.SettingsTwoFactorDisable, controllers.DisableTwoFactorEndpoint)
	}
}
