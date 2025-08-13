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
		notifications.DELETE("/:id", controllers.DeleteNotificationEndpoint)
		notifications.GET("/unread-count", controllers.GetUnreadNotificationCountEndpoint)
		notifications.GET("/recent", controllers.GetRecentNotificationsEndpoint)
	}
}
