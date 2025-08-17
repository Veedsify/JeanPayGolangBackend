package routes

import (
	"log"

	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/middlewares"
	"github.com/Veedsify/JeanPayGoBackend/routes/endpoints"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {

	v1 := router.Group(constants.AdminBase)
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	v1.Use(middlewares.AuthMiddleware(jwtService))
	{
		endpoints.AdminRoutes(v1)
	}
}
