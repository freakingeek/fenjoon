package handlers

import (
	"net/http"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
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

	pushToken := models.PushToken{Token: request.Token, UserID: userId}
	if err := database.DB.Where("token = ?", request.Token).FirstOrCreate(&pushToken).Error; err != nil {
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
