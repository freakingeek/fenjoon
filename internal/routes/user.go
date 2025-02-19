package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/users")

	v1.GET("/me", handlers.GetUserById)
	v1.PATCH("/me", handlers.UpdateUserById)
	v1.GET("/me/stories", handlers.GetUserStories)
}
