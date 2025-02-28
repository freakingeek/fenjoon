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
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var request struct {
		Text string `json:"text" binding:"required,min=25,max=250"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryCharLimit, Data: nil})
		return
	}

	story := models.Story{Text: request.Text, UserID: userId}

	if err := database.DB.Create(&story).Preload("User").First(&story, story.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.StoryNotCreated, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryCreated, Data: story})
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
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("User").Order("id DESC").Limit(limit).Offset(offset).Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	for i := range stories {
		var likeCount int64
		database.DB.Model(&models.Like{}).Where("story_id = ?", stories[i].ID).Count(&likeCount)
		stories[i].LikesCount = uint(likeCount)
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

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	if err := database.DB.Preload("User").Where("id = ?", storyId).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var likesCount int64
	if err := database.DB.Model(&models.Like{}).Where("story_id = ?", storyId).Count(&likesCount).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	story.LikesCount = uint(likesCount)

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: story})
}

func UpdateStory(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var request struct {
		Text string `json:"text" binding:"required,min=25,max=250"`
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryCharLimit, Data: nil})
		return
	}

	var story models.Story

	if err := database.DB.Preload("User").Where("id = ? AND user_id = ?", storyId, userId).First(&story).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	story.Text = request.Text

	if err := database.DB.Save(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryEdited, Data: story})
}

func DeleteStory(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var story models.Story

	if err := database.DB.Preload("User").Where("id = ? AND user_id = ?", storyId, userId).First(&story).Error; err != nil {
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
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	var existingLike models.Like
	if err := database.DB.Where("story_id = ? AND user_id = ?", storyId, userId).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusConflict, responses.ApiResponse{Status: http.StatusConflict, Message: messages.StoryAlreadyLiked, Data: nil})
		return
	}

	like := models.Like{StoryID: uint(storyId), UserID: userId}
	if err := database.DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryLiked, Data: true})
}

func DislikeStoryById(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var like models.Like
	if err := database.DB.Where("story_id = ? AND user_id = ?", storyId, userId).First(&like).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryDisliked, Data: false})
}

func GetStoryLikers(c *gin.Context) {
	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, storyId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var users []models.User
	if err := database.DB.
		Joins("JOIN likes ON likes.user_id = users.id").
		Where("likes.story_id = ?", storyId).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: users})
}
