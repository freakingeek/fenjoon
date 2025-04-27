package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/users")

	v1.GET("/me", handlers.GetCurrentUser)
	v1.PATCH("/me", handlers.UpdateCurrentUser)
	v1.GET("/me/stories", handlers.GetCurrentUserStories) // All User stories (public + private)
	v1.GET("/me/private-story-count", handlers.GetUserPrivateStoriesCount)

	v1.GET(":id", handlers.GetUserById)
	v1.GET(":id/stories", handlers.GetUserPublicStories) // Public Stories
	v1.GET(":id/comments", handlers.GetUserComments) // Public Comments
}
