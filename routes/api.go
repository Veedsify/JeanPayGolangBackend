package routes

import (
	"log"

	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/middlewares"
	"github.com/Veedsify/JeanPayGoBackend/routes/endpoints"
	"github.com/gin-gonic/gin"
)

func ApiRoutes(router *gin.Engine) {

	v1 := router.Group(constants.APIBase)
	public := v1.Group("/")
	{
		endpoints.AuthRoutes(public)
	}
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	protected := v1.Group(constants.ProtectedBase)
	protected.Use(middlewares.AuthMiddleware(jwtService))
	{
		endpoints.UserRoutes(protected)
		endpoints.WalletRoutes(protected)
		endpoints.ConvertRoutes(protected)
		endpoints.TransactionRoutes(protected)
		endpoints.NotificationRoutes(protected)
		endpoints.SettingsRoutes(protected)
		endpoints.DashboardRoutes(protected)
	}

	// Seed routes (public for easy access during development)
	seed := v1.Group("/")
	{
		endpoints.SeedRoutes(seed)
	}
}
