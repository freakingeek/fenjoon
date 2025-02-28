package handlers

import (
	"net/http"
	"strconv"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
)

func CreateStory(c *gin.Context) {
	token, err := auth.ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"story": nil},
		})
		return
	}

	claims, err := auth.ParseJWTToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"story": nil},
		})
		return
	}

	floatUserId, ok := claims["id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"story": nil},
		})
		return
	}

	var request struct {
		Text string `json:"text" binding:"required,min=25,max=256"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotCreated, Data: map[string]interface{}{"story": nil}})
		return
	}

	story := models.Story{Text: request.Text, UserID: uint(floatUserId)}

	if err := database.DB.Create(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.StoryNotCreated, Data: map[string]interface{}{"story": nil}})
		return
	}

	database.DB.Preload("User").First(&story, story.ID)

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryCreated, Data: map[string]interface{}{"story": story}})
}

func GetAllStories(c *gin.Context) {
	var stories []models.Story
	var total int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	if err := database.DB.Model(&models.Story{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	if err := database.DB.Preload("User").Order("id DESC").Limit(limit).Offset(offset).Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"stories": stories,
			"pagination": map[string]interface{}{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}

func GetStoryById(c *gin.Context) {
	var story models.Story

	id := c.Param("id")

	if err := database.DB.Where("id = ?", id).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: map[string]interface{}{"story": nil}})
		return
	}

	database.DB.Preload("User").First(&story, story.ID)

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

func LikeStoryById(c *gin.Context) {
	token, err := auth.ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    nil,
		})
		return
	}

	claims, err := auth.ParseJWTToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    nil,
		})
		return
	}

	floatUserId, ok := claims["id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	userId := uint(floatUserId)

	storyIdParam := c.Param("id")
	storyId, err := strconv.ParseUint(storyIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.UserNotFound,
			Data:    nil,
		})
		return
	}

	var existingLike models.Like
	if err := database.DB.Where("story_id = ? AND user_id = ?", storyId, userId).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusConflict, responses.ApiResponse{
			Status:  http.StatusConflict,
			Message: messages.StoryAlreadyLiked,
			Data:    nil,
		})
		return
	}

	like := models.Like{
		StoryId: uint(storyId),
		UserId:  userId,
	}

	if err := database.DB.Create(&like).Error; err != nil {
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
		Data:    gin.H{"liked": true},
	})
}

func DislikeStoryById(c *gin.Context) {
	token, err := auth.ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    nil,
		})
		return
	}

	claims, err := auth.ParseJWTToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    nil,
		})
		return
	}

	floatUserId, ok := claims["id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
		return
	}

	userId := uint(floatUserId)

	storyIdParam := c.Param("id")
	storyId, err := strconv.ParseUint(storyIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var like models.Like
	if err := database.DB.Where("story_id = ? AND user_id = ?", storyId, userId).First(&like).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.GeneralAccessDenied,
			Data:    nil,
		})
		return
	}

	if err := database.DB.Delete(&like).Error; err != nil {
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
		Data:    gin.H{"liked": false},
	})
}

func GetStoryLikers(c *gin.Context) {
	storyIdParam := c.Param("id")
	storyId, err := strconv.ParseUint(storyIdParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.StoryNotFound,
			Data:    nil,
		})
		return
	}

	var users []models.User
	if err := database.DB.
		Joins("JOIN likes ON likes.user_id = users.id").
		Where("likes.story_id = ?", storyId).
		Find(&users).Error; err != nil {
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
		Data:    users,
	})
}
