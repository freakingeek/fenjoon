package routes

import "github.com/gin-gonic/gin"

func StoriesRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/stories")

	v1.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "stories",
		})
	})
}
