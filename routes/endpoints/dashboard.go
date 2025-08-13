package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func DashboardRoutes(router *gin.RouterGroup) {
	dashboard := router.Group(constants.DashboardBase)
	{
		dashboard.GET(constants.DashboardOverview, controllers.GetDashboardOverviewEndpoint)
		dashboard.GET(constants.DashboardStats, controllers.GetDashboardStatsEndpoint)
		dashboard.GET("/summary", controllers.GetDashboardSummaryEndpoint)
		dashboard.GET("/recent-activity", controllers.GetRecentActivityEndpoint)
		dashboard.GET("/wallet-overview", controllers.GetWalletOverviewEndpoint)
		dashboard.GET("/conversion-stats", controllers.GetConversionStatsEndpoint)
		dashboard.GET("/monthly-stats", controllers.GetMonthlyStatsEndpoint)
		dashboard.GET("/transaction-trends", controllers.GetTransactionTrendsEndpoint)
		dashboard.GET("/transaction-stats", controllers.GetTransactionStatsEndpoint)
		dashboard.GET("/charts-data", controllers.GetDashboardChartsDataEndpoint)
	}
}
