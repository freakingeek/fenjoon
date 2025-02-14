package handlers

import (
	"net/http"

	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
)

func CreateStory(c *gin.Context) {
	var request struct {
		Text string `json:"text" binding:"required,min=25,max=256"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotCreated, Data: map[string]interface{}{"story": nil}})
		return
	}

	story := models.Story{Text: request.Text}

	if err := database.DB.Create(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.StoryNotCreated, Data: map[string]interface{}{"story": story}})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryCreated, Data: map[string]interface{}{"story": story}})
}

func GetAllStories(c *gin.Context) {
	var stories []models.Story

	if err := database.DB.Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: map[string]interface{}{"stories": nil}})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: map[string]interface{}{"stories": stories}})
}

func GetStoryById(c *gin.Context) {
	var story models.Story

	id := c.Param("id")

	if err := database.DB.Where("id = ?", id).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: map[string]interface{}{"story": nil}})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: map[string]interface{}{"story": story}})
}

func UpdateStory(c *gin.Context) {
	var request struct {
		Text string `json:"text" binding:"required,min=25,max=256"`
	}

	id := c.Param("id")

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralFailed, Data: map[string]interface{}{"story": nil}})
		return
	}

	var story models.Story

	if err := database.DB.Where("id = ?", id).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: map[string]interface{}{"story": nil}})
		return
	}

	story.Text = request.Text

	if err := database.DB.Save(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: map[string]interface{}{"story": nil}})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryEdited, Data: map[string]interface{}{"story": story}})
}

func DeleteStory(c *gin.Context) {
	id := c.Param("id")

	var story models.Story

	if err := database.DB.Where("id = ?", id).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: map[string]interface{}{"story": nil}})
		return
	}

	if err := database.DB.Delete(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: map[string]interface{}{"story": story}})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryDeleted, Data: map[string]interface{}{"story": story}})
}
