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
		dashboard.GET(constants.DashboardSummary, controllers.GetDashboardSummaryEndpoint)
		dashboard.GET(constants.DashboardRecentActivity, controllers.GetRecentActivityEndpoint)
		dashboard.GET(constants.DashboardWalletOverview, controllers.GetWalletOverviewEndpoint)
		dashboard.GET(constants.DashboardConversionStats, controllers.GetConversionStatsEndpoint)
		dashboard.GET(constants.DashboardMonthlyStats, controllers.GetMonthlyStatsEndpoint)
		dashboard.GET(constants.DashboardTransactionTrends, controllers.GetTransactionTrendsEndpoint)
		dashboard.GET(constants.DashboardTransactionStats, controllers.GetTransactionStatsEndpoint)
		dashboard.GET(constants.DashboardChartsData, controllers.GetDashboardChartsDataEndpoint)
	}
}
