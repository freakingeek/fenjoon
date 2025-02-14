package routes

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")

	UserRoutes(v1)
	AuthRoutes(v1)
	StoryRoutes(v1)
}
