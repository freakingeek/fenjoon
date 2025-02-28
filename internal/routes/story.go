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

	v1.GET(":id/likes", handlers.GetStoryLikers)
	v1.POST(":id/likes", handlers.LikeStoryById)
	v1.DELETE(":id/likes", handlers.DislikeStoryById)

	v1.GET(":id/comments", handlers.GetStoryComments)
	v1.POST(":id/comments", handlers.CommentStoryById)
	v1.DELETE(":id/comments/:commentId", handlers.UncommentStoryById)

	v1.POST(":id/shares", handlers.ShareStoryById)
}
