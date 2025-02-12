package routes

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")

	StoriesRoutes(v1)
}
