package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/notifications")

	v1.GET("", handlers.GetUserNotifications)
	v1.GET(":id", handlers.GetNotificationById)
	v1.GET("/unread-count", handlers.GetUserNotificationsUnreadCount)
	v1.PATCH("/read", handlers.MarkNotificationsAsRead)
}
