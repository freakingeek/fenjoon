package routes

import (
	"github.com/freakingeek/fenjoon/internal/handlers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.RouterGroup) {
	v1 := r.Group("/auth")

	v1.POST("/otp/send", handlers.SendOTPHandler)
	v1.POST("/otp/verify", handlers.VerifyOTPHandler)
}
