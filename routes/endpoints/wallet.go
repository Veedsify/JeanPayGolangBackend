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
		wallet.POST(constants.WalletWithdrawFee, controllers.CalculateWithdrawalFeeEndpoint)
		wallet.POST(constants.WalletWithdrawValidate, controllers.ValidateWithdrawalEndpoint)
		wallet.GET(constants.WalletWithdrawDetails, controllers.GetWithdrawalDetailsEndpoint)
		wallet.GET(constants.WalletWithdrawHistory, controllers.GetUserWithdrawalsEndpoint)
		wallet.GET(constants.WalletHistory, controllers.GetWalletHistoryEndpoint)
		wallet.POST(constants.WalletUpdateAfterPayment, controllers.UpdateWalletAfterPaymentEndpoint)
	}
}
