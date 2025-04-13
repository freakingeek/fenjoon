package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
)

func isFarsiText(text string) bool {
	// Updated regex to match Persian characters and common punctuation, including the Persian comma (،)
	re := regexp.MustCompile(`^[\p{Arabic}\s\x{200C}\x{0640}.,()\[\]؟!؛:،]+$`)
	return re.MatchString(text)
}

func GetCurrentUser(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"nickname":  user.Nickname,
			"phone":     user.Phone,
			"bio":       user.Bio,
		},
	})
}

func UpdateCurrentUser(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var request struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Nickname  string `json:"nickname"`
		Bio       string `json:"bio" binding:"max=150"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if (request.FirstName != "" && !isFarsiText(request.FirstName)) ||
		(request.LastName != "" && !isFarsiText(request.LastName)) ||
		(request.Nickname != "" && !isFarsiText(request.Nickname)) ||
		(request.Bio != "" && !isFarsiText(request.Bio)) {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.UserForbiddenName, Data: nil})
		return
	}

	var user models.User

	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	updates := map[string]any{}
	updates["first_name"] = strings.TrimSpace(request.FirstName)
	updates["last_name"] = strings.TrimSpace(request.LastName)
	updates["nickname"] = strings.TrimSpace(request.Nickname)
	updates["bio"] = strings.TrimSpace(request.Bio)

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.UserEdited, Data: user})
}

func GetUserById(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"nickname":  user.Nickname,
			"bio":       user.Bio,
		},
	})
}

func GetUserStories(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var stories []models.Story
	var total int64

	if err := database.DB.Model(&models.Story{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("User").Order("id DESC").Where("user_id = ?", userId).Limit(limit).Offset(offset).Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	for i := range stories {
		var likesCount int64
		if err := database.DB.Model(&models.Like{}).Where("story_id = ?", stories[i].ID).Count(&likesCount).Error; err != nil {
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
		stories[i].CommentsCount = uint(commentsCount)
		stories[i].IsLikedByUser = isLikedByUser
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"stories": stories,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)), // Calculate total pages
			},
		},
	})
}

func GetUserComments(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var comments []models.Comment
	var total int64

	if err := database.DB.Model(&models.Comment{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("User").Order("id DESC").Where("user_id = ?", userId).Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	for i := range comments {
		var likesCount int64
		if err := database.DB.Model(&models.CommentLike{}).Where("comment_id = ?", comments[i].ID).Count(&likesCount).Error; err != nil {
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
		comments[i].IsEditableByUser = uint(userId) == comments[i].UserID
		comments[i].IsDeletableByUser = uint(userId) == comments[i].UserID
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"comments": comments,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)), // Calculate total pages
			},
		},
	})
}
