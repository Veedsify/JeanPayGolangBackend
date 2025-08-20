package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func TransactionRoutes(router *gin.RouterGroup) {
	transactions := router.Group(constants.TransactionsBase)
	{
		transactions.POST(constants.TransactionsNew, controllers.CreateTransactionEndpoint)
		transactions.GET(constants.TransactionsAll, controllers.GetAllTransactionsEndpoint)
		transactions.GET(constants.TransactionsUserHistory, controllers.GetUserTransactionHistoryEndpoint)
		transactions.GET(constants.TransactionsDetails, controllers.GetTransactionDetailsEndpoint)
		transactions.PUT(constants.TransactionsUpdateStatus, controllers.UpdateTransactionStatusEndpoint)
		transactions.GET(constants.TransactionsStats, controllers.GetTransactionStatsEndpoint)
		transactions.POST(constants.TransactionsFilter, controllers.FilterTransactionsEndpoint)
	}
}
