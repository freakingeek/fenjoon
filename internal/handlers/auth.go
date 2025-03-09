package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SendOTP(c *gin.Context) {
	var request struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	existingTTL, err := database.RedisClient.TTL(context.Background(), "otp:"+request.Phone).Result()
	if err == nil && existingTTL > 0 {
		c.JSON(http.StatusTooManyRequests, responses.ApiResponse{
			Status:  http.StatusTooManyRequests,
			Message: fmt.Sprintf(messages.OTPTryAgain, int(existingTTL.Seconds())),
			Data:    nil,
		})
		return
	}

	// Generate a 5-digit OTP
	rand.NewSource(time.Now().UnixNano())
	otp := rand.Intn(90000) + 10000

	err = database.RedisClient.Set(context.Background(), "otp:"+request.Phone, otp, 1*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	go services.SendOTPViaSMS(request.Phone, otp)

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]string{
			"status": "success",
			"phone":  request.Phone,
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
			Data:    nil,
		})
		return
	}

	storedOTP, err := database.RedisClient.Get(context.Background(), "otp:"+request.Phone).Result()
	if err != nil || storedOTP != request.Code {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.OTPInvalid,
			Data:    nil,
		})
		return
	}

	database.RedisClient.Del(context.Background(), "otp:"+request.Phone)

	isNewUser := false

	var user models.User
	if err := database.DB.Where("phone = ?", request.Phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				FirstName: "",
				LastName:  "",
				Nickname:  "",
				Phone:     request.Phone,
			}

			if err := database.DB.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, responses.ApiResponse{
					Status:  http.StatusInternalServerError,
					Message: messages.GeneralFailed,
					Data:    nil,
				})
				return
			}

			isNewUser = true
		} else {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{
				Status:  http.StatusInternalServerError,
				Message: messages.GeneralFailed,
				Data:    nil,
			})
			return
		}
	}

	accessToken, err := auth.GenerateJWTToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	refreshToken := uuid.New().String()
	err = database.RedisClient.Set(context.Background(), "refresh:"+refreshToken, user.ID, 30*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"status":    "success",
			"isNewUser": isNewUser,
			"tokens": map[string]interface{}{
				"accessToken":  accessToken,
				"refreshToken": refreshToken,
			},
		},
	})
}

func RefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	userIDStr, err := database.RedisClient.Get(context.Background(), "refresh:"+request.RefreshToken).Result()
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.InvalidRefreshToken,
			Data:    nil,
		})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	newAccessToken, err := auth.GenerateJWTToken(uint(userIDUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"status": "success",
			"tokens": map[string]interface{}{
				"accessToken":  newAccessToken,
				"refreshToken": request.RefreshToken,
			},
		},
	})
}
