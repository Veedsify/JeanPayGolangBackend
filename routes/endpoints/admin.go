package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.RouterGroup) {
	admin := router.Group(constants.AdminBase)
	{
		// Admin dashboard routes
		admin.POST(constants.AdminDashboard, controllers.GetAdminDashboardStatistics)
		// Admin user management routes
		admin.POST(constants.AdminUsersAll, controllers.GetAdminUsersAll)
		admin.POST(constants.AdminUsersDetails, controllers.AdminUsersDetails)
		// Admin transaction management routes
		admin.POST(constants.AdminTransactionsBase+constants.AdminTransactionsAll, controllers.GetAdminTransactionsAll)
		admin.POST(constants.AdminTransactionsBase+constants.AdminTransactionsDetails, controllers.GetAdminTransactionDetails)
		admin.PATCH(constants.AdminTransactionsBase+constants.AdminTransactionsApprove, controllers.ApproveAdminTransaction)
		admin.PATCH(constants.AdminTransactionsBase+constants.AdminTransactionsReject, controllers.RejectAdminTransaction)
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsStatus, controllers.AdminTransactionStatus)
		admin.POST(constants.AdminTransactionsBase+constants.AdminTransactionsOverview, controllers.AdminTransactionsOverview)
		// Additional rates endpoints can be added here
		admin.GET(constants.AdminRatesBase+constants.AdminRatesHistory, controllers.AdminRatesHistory)
		admin.POST(constants.AdminRatesBase+constants.AdminRatesAdd, controllers.AdminRatesAdd)
	}
}
