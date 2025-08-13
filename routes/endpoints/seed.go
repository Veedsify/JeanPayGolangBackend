package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func SeedRoutes(router *gin.RouterGroup) {
	seed := router.Group("/seed")
	{
		// Exchange rate seeding
		seed.POST("/exchange-rates", controllers.SeedExchangeRatesEndpoint)

		// All default data seeding
		seed.POST("/default-data", controllers.SeedAllDefaultDataEndpoint)

		// Test data seeding (for development)
		seed.POST("/test-data", controllers.SeedTestDataEndpoint)
	}
}
