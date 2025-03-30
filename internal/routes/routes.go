package routes

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")

	UserRoutes(v1)
	AuthRoutes(v1)
	PushRoutes(v1)
	StoryRoutes(v1)
	CommentRoutes(v1)
	NotificationRoutes(v1)
}
