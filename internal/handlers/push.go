package handlers

import (
	"errors"
	"net/http"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPushToken(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		userId = 0
	}

	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var pushToken models.PushToken
	err = database.DB.Where("token = ?", request.Token).First(&pushToken).Error

	if err == nil {
		if pushToken.UserID == 0 && userId != 0 {
			pushToken.UserID = userId
			if err := database.DB.Save(&pushToken).Error; err != nil {
				c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
				return
			}
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		var existingToken models.PushToken
		err = database.DB.Where("user_id = ?", userId).First(&existingToken).Error

		if err == nil {
			existingToken.Token = request.Token
			if err := database.DB.Save(&existingToken).Error; err != nil {
				c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
				return
			}
		} else {
			// NOTE: Push token doesn't exist; create a new one
			newPushToken := models.PushToken{UserID: userId, Token: request.Token}
			if err := database.DB.Create(&newPushToken).Error; err != nil {
				c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
				return
			}
		}
	} else {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: pushToken})
}

func UnregisterPushToken(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		userId = 0
	}

	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var pushToken models.PushToken
	if err := database.DB.Where("token = ? AND user_id = ?", request.Token, userId).First(&pushToken).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.GeneralNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&pushToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: pushToken})
}
