package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(admin *gin.RouterGroup) {
	{
		// Admin dashboard routes
		admin.POST(constants.AdminDashboard, controllers.GetAdminDashboardStatistics)

		// Admin user management routes
		admin.POST(constants.AdminUsersBase+constants.AdminUsersAll, controllers.GetAdminUsersAll)
		admin.POST(constants.AdminUsersBase+constants.AdminUsersDetails, controllers.AdminUsersDetails)
		admin.PATCH(constants.AdminUsersBase+constants.AdminUserUpdate, controllers.AdminUserUpdate)
		admin.PATCH(constants.AdminUsersBase+constants.AdminUsersBlock, controllers.BlockUser)
		admin.PATCH(constants.AdminUsersBase+constants.AdminUsersUnblock, controllers.UnblockUser)
		admin.GET(constants.AdminUsersBase+constants.AdminUsersTransactions, controllers.GetUserTransactions)
		admin.GET(constants.AdminUsersBase+constants.AdminUsersWallet, controllers.GetUserWallet)
		admin.GET(constants.AdminUsersBase+constants.AdminUsersActivityLogs, controllers.GetUserActivityLogs)
		admin.POST(constants.AdminUsersBase+constants.AdminUsersSearch, controllers.SearchUsers)

		// Admin transaction management routes
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsAll, controllers.GetAdminTransactionsAll)
		admin.POST(constants.AdminTransactionsBase+constants.AdminTransactionsDetails, controllers.GetAdminTransactionDetails)
		admin.PATCH(constants.AdminTransactionsBase+constants.AdminTransactionsApprove, controllers.ApproveAdminTransaction)
		admin.PATCH(constants.AdminTransactionsBase+constants.AdminTransactionsReject, controllers.RejectAdminTransaction)
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsStatus, controllers.AdminTransactionStatus)
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsOverview, controllers.AdminTransactionsOverview)
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsPending, controllers.GetPendingTransactions)
		admin.GET(constants.AdminTransactionsBase+constants.AdminTransactionsFailed, controllers.GetFailedTransactions)
		admin.POST(constants.AdminTransactionsBase+constants.AdminTransactionsNotes, controllers.AddTransactionNote)

		// Admin rates management routes
		admin.GET(constants.AdminRatesBase+constants.AdminRatesHistory, controllers.AdminRatesHistory)
		admin.POST(constants.AdminRatesBase+constants.AdminRatesAdd, controllers.AdminRatesAdd)
		admin.PATCH(constants.AdminRatesBase+constants.AdminRatesUpdateById, controllers.UpdateRate)
		admin.PATCH(constants.AdminRatesBase+constants.AdminRatesToggle, controllers.ToggleRateStatus)
		admin.DELETE(constants.AdminRatesBase+constants.AdminRatesDelete, controllers.DeleteRate)
	}
}
