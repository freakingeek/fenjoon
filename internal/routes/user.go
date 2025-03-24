package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/users")

	v1.GET("/me", handlers.GetCurrentUser)
	v1.PATCH("/me", handlers.UpdateCurrentUser)

	v1.GET(":id", handlers.GetUserById)
	v1.GET(":id/stories", handlers.GetUserStories)
}
