package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/gin-gonic/gin"
)

func SendOTP(c *gin.Context) {
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

	existingTTL, err := database.RedisClient.TTL(context.Background(), "otp:"+request.Phone).Result()
	if err == nil && existingTTL > 0 {
		c.JSON(http.StatusTooManyRequests, responses.ApiResponse{
			Status:  http.StatusTooManyRequests,
			Message: fmt.Sprintf(messages.OTPTryAgain, int(existingTTL.Seconds())),
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	// Generate a 5-digit OTP
	rand.NewSource(time.Now().UnixNano())
	otp := rand.Intn(90000) + 10000

	err = database.RedisClient.Set(context.Background(), "otp:"+request.Phone, otp, 30*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	go services.SendOTPViaSMS(request.Phone, otp)

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]string{
			"status": "success",
		},
	})
}

func VerifyOTP(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	storedOTP, err := database.RedisClient.Get(context.Background(), "otp:"+request.Phone).Result()
	if storedOTP != request.Code || err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.OTPInvalid,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	database.RedisClient.Del(context.Background(), "otp:"+request.Phone)

	user := models.User{FirstName: "", LastName: "", Nickname: ""}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	token, err := auth.GenerateJWTToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"status": "failed"},
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"status": "success",
			"tokens": map[string]interface{}{
				"accessToken": token,
			},
		},
	})
}
