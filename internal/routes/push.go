package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func PushRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/push")

	v1.POST("/register", handlers.RegisterPushToken)
	v1.DELETE("/unregister", handlers.UnregisterPushToken)
}
