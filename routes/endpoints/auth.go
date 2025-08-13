package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup) {
	{
		auth := router.Group(constants.AuthBase)
		auth.POST(constants.AuthSignup, controllers.RegisterUserEndpoint)
		auth.POST(constants.AuthLogin, controllers.LoginUserEndpoint)
		auth.POST(constants.AuthVerify, controllers.VerifyUserEndpoint)
		auth.POST(constants.AuthPasswordReset, controllers.PasswordResetEndpoint)
		auth.GET(constants.AuthResetPassWordVerify, controllers.ResetPasswordTokenVerifyEndpoint)
		auth.POST(constants.AuthResetPassword, controllers.ResetPasswordEndpoint)
	}
}
