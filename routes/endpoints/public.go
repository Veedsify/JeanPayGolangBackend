package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func PublicRoutes(router *gin.RouterGroup) {
	{
		router.GET(constants.ConvertBase+constants.ConvertRates, controllers.GetExchangeRatesEndpoint)
		router.GET(constants.SettingsBase+constants.SettingsPlatform, controllers.GetPlatformSettingsEndpoint)
		router.GET(constants.AdminPaymentAccountsBase+constants.AdminPaymentAccountsActive, controllers.GetActivePaymentAccounts)
		router.POST(constants.ContactRoute, controllers.ContactInformation)
	}
}
