package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func ConvertRoutes(router *gin.RouterGroup) {
	convert := router.Group(constants.ConvertBase)
	{
		convert.GET(constants.ConvertRates, controllers.GetExchangeRatesEndpoint)
		convert.POST(constants.ConvertCalculate, controllers.CalculateConversionEndpoint)
		convert.POST(constants.ConvertExchange, controllers.ExecuteConversionEndpoint)
		convert.GET("/history", controllers.GetConversionHistoryEndpoint)
	}
}
