package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/auth")

	v1.POST("/otp/send", handlers.SendOTP)
	v1.POST("/otp/verify", handlers.VerifyOTP)

	v1.POST("/refresh", handlers.RefreshToken)
}
