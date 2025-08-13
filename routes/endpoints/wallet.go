package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func WalletRoutes(router *gin.RouterGroup) {
	wallet := router.Group(constants.WalletBase)
	{
		wallet.GET(constants.WalletBalance, controllers.GetWalletBalanceEndpoint)
		wallet.POST(constants.WalletTopUp, controllers.TopUpWalletEndpoint)
		wallet.GET(constants.WalletTopUpDetails, controllers.GetTopUpDetailsEndpoint)
		wallet.POST(constants.WalletWithdraw, controllers.WithdrawFromWalletEndpoint)
		wallet.GET(constants.WalletHistory, controllers.GetWalletHistoryEndpoint)
		wallet.POST(constants.WalletUpdateAfterPayment, controllers.UpdateWalletAfterPaymentEndpoint)
	}
}
