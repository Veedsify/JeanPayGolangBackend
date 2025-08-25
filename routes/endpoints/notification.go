package endpoints

import (
	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/controllers"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(router *gin.RouterGroup) {
	notifications := router.Group(constants.NotificationsBase)
	{
		notifications.GET(constants.NotificationsAll, controllers.GetAllNotificationsEndpoint)
		notifications.PUT(constants.NotificationsMarkRead, controllers.MarkNotificationReadEndpoint)
		notifications.PUT(constants.NotificationsMarkAllRead, controllers.MarkAllNotificationsReadEndpoint)
		notifications.POST(constants.NotificationMarkReadBulk, controllers.NotificationMarkReadBulkEndpoint)
		notifications.DELETE(constants.NotificationDeleteBulk, controllers.DeleteNotificationEndpoint)
	}
}
