package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/freakingeek/fenjoon/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	userId, _ := auth.GetUserIdFromContext(c)

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
		var likesCount int64
		if err := database.DB.Model(&models.Like{}).Where("story_id = ?", stories[i].ID).Count(&likesCount).Error; err != nil {
			c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
			return
		}

		var sharesCount int64
		if err := database.DB.Model(&models.Share{}).Where("story_id = ?", stories[i].ID).Count(&sharesCount).Error; err != nil {
			c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
			return
		}

		var commentsCount int64
		if err := database.DB.Model(&models.Comment{}).Where("story_id = ?", stories[i].ID).Count(&commentsCount).Error; err != nil {
			c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
			return
		}

		var isLikedByUser bool
		if err := database.DB.Model(&models.Like{}).
			Where("story_id = ? AND user_id = ?", stories[i].ID, userId).
			Select("COUNT(*) > 0").
			Find(&isLikedByUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		stories[i].LikesCount = uint(likesCount)
		stories[i].SharesCount = uint(sharesCount)
		stories[i].CommentsCount = uint(commentsCount)
		stories[i].IsLikedByUser = isLikedByUser
		stories[i].IsEditableByUser = userId == stories[i].UserID
		stories[i].IsDeletableByUser = userId == stories[i].UserID
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

	userId, _ := auth.GetUserIdFromContext(c)

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

	var sharesCount int64
	if err := database.DB.Model(&models.Share{}).Where("story_id = ?", storyId).Count(&sharesCount).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var commentsCount int64
	if err := database.DB.Model(&models.Comment{}).Where("story_id = ?", storyId).Count(&commentsCount).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var isLikedByUser bool
	if err := database.DB.Model(&models.Like{}).
		Where("story_id = ? AND user_id = ?", storyId, userId).
		Select("COUNT(*) > 0").
		Find(&isLikedByUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	story.LikesCount = uint(likesCount)
	story.SharesCount = uint(sharesCount)
	story.CommentsCount = uint(commentsCount)
	story.IsLikedByUser = isLikedByUser
	story.IsEditableByUser = userId == story.UserID
	story.IsDeletableByUser = userId == story.UserID

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
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.StoryNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryDeleted, Data: story})
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

	var pushToken models.PushToken
	if err := database.DB.Where("user_id = ?", story.UserID).First(&pushToken).Error; err == nil {
		text := fmt.Sprintf("%s از داستانت خوشش اومد", utils.GetUserDisplayName(user))

		if err := services.SendPushNotification([]string{pushToken.Token}, text); err != nil {
			fmt.Printf("Failed to send push notification: %v\n", err)
		}
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

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.StoryDisliked, Data: true})
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
		Where("likes.deleted_at IS NULL").
		Select("DISTINCT users.*").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: users})
}

func IsStoryLikedByUser(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var like models.Like
	if err = database.DB.Where("user_id = ? AND story_id = ?", userId, storyId).First(&like).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	isLiked := err != gorm.ErrRecordNotFound

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: isLiked})
}

func CommentStoryById(c *gin.Context) {
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

	var request struct {
		Text string `json:"text" binding:"required,min=5,max=250"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	comment := models.Comment{StoryID: uint(storyId), UserID: uint(userId), Text: request.Text}
	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    nil,
		})
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

	var pushToken models.PushToken
	if err := database.DB.Where("user_id = ?", story.UserID).First(&pushToken).Error; err == nil {
		text := fmt.Sprintf("%s نقد جدیدی روی داستانت ثبت کرد", utils.GetUserDisplayName(user))

		if err := services.SendPushNotification([]string{pushToken.Token}, text); err != nil {
			fmt.Printf("Failed to send push notification: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: comment})
}

func GetStoryComments(c *gin.Context) {
	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var comments []models.Comment
	var total int64

	userId, _ := auth.GetUserIdFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	if err := database.DB.Model(&comments).Where("story_id = ?", storyId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Where("story_id = ?", storyId).Preload("User").Order("id DESC").Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: "Failed to fetch comments", Data: nil})
		return
	}

	for i := range comments {
		var likesCount int64
		if err := database.DB.Model(&models.CommentLike{}).Where("story_id = ?", comments[i].ID).Count(&likesCount).Error; err != nil {
			c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
			return
		}

		var isLikedByUser bool
		if err := database.DB.Model(&models.CommentLike{}).
			Where("comment_id = ? AND user_id = ?", comments[i].ID, userId).
			Select("COUNT(*) > 0").
			Find(&isLikedByUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		comments[i].LikesCount = uint(likesCount)
		comments[i].IsLikedByUser = isLikedByUser
		comments[i].IsEditableByUser = userId == comments[i].UserID
		comments[i].IsDeletableByUser = userId == comments[i].UserID
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"comments": comments,
			"pagination": map[string]interface{}{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}

func ShareStoryById(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		// Ignore share request if user is not logged-in
		c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: nil})
		return
	}

	storyId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var lastShare models.Share
	err = database.DB.Where("user_id = ? AND story_id = ?", userId, storyId).Order("created_at DESC").First(&lastShare).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err == gorm.ErrRecordNotFound || time.Since(lastShare.CreatedAt) >= 5*time.Minute {
		share := models.Share{UserID: uint(userId), StoryID: uint(storyId)}

		if err := database.DB.Create(&share).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: true})
		return
	}

	c.JSON(http.StatusTooManyRequests, responses.ApiResponse{Status: http.StatusTooManyRequests, Message: messages.StoryShareLimit, Data: nil})
}
