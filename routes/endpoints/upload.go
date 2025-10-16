package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func UploadRoutes(router *gin.RouterGroup) {
	upload := router.Group("/upload")
	{
		upload.POST("/receipt", controllers.UploadReceiptEndpoint)
		upload.GET("/receipt/:transaction_id", controllers.GetReceiptEndpoint)
	}
}