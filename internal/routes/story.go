package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func StoryRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/stories")

	v1.POST("", handlers.CreateStory)
	v1.GET("", handlers.GetAllStories)
	v1.GET(":id", handlers.GetStoryById)
	v1.PUT(":id", handlers.UpdateStory)
	v1.DELETE(":id", handlers.DeleteStory)
}
