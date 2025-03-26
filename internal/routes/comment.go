package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func CommentRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/comments")

	v1.GET(":id/likes", handlers.GetCommentLikers)
	v1.POST(":id/likes", handlers.LikeCommentById)
	v1.DELETE(":id/likes", handlers.DislikeCommentById)
}
