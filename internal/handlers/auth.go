package handlers

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/gin-gonic/gin"
)

func SendOTPHandler(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	// Generate a 5-digit OTP
	rand.NewSource(time.Now().UnixNano())
	otp := rand.Intn(90000) + 10000

	err := database.RedisClient.Set(context.Background(), "otp:"+request.Phone, otp, 5*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
		})
		return
	}

	go services.SendOTPViaSMS(request.Phone, otp)

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]bool{
			"status": true,
		},
	})
}
