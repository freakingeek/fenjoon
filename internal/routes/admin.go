package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/admin")

	v1.GET("/story-reports", handlers.GetStoryReports)
	v1.GET("/story-reports/:id", handlers.GetStoryReport)
	v1.PUT("/story-reports/:id/resolve", handlers.ResolveStoryReport)
	v1.PUT("/story-reports/:id/reject", handlers.RejectStoryReport)
}
