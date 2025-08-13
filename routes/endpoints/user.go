package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup) {
	{
		auth := router.Group(constants.UserBase)
		auth.POST(constants.UserRetrieve, controllers.FetchUserEndpoint)
	}
}
