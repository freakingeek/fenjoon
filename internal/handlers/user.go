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
	userId, _ := auth.GetUserIdFromContext(c)

	targetUserId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.First(&user, targetUserId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&models.Story{}).Where("user_id = ?", targetUserId)

	if uint(targetUserId) != userId {
		query = query.Where("is_private = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var stories []models.Story
	if err := query.Preload("User").Order("id DESC").Limit(limit).Offset(offset).Find(&stories).Error; err != nil {
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
		stories[i].IsEditableByUser = userId == stories[i].UserID
		stories[i].IsDeletableByUser = userId == stories[i].UserID
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

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	var comments []models.Comment
	var total int64

	if err := database.DB.Model(&models.Comment{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("User").Preload("Story.User").Order("id DESC").Where("user_id = ?", userId).Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
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

		var storyLikesCount int64
		if err := database.DB.Model(&models.Story{}).Where("id = ?", comments[i].StoryID).Count(&storyLikesCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		var storyCommentsCount int64
		if err := database.DB.Model(&models.Comment{}).Where("story_id = ?", comments[i].StoryID).Count(&storyCommentsCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		var isStoryLikedByUser bool
		if err := database.DB.Model(&models.Story{}).
			Where("id = ? AND user_id = ?", comments[i].StoryID, userId).
			Select("COUNT(*) > 0").
			Find(&isStoryLikedByUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
			return
		}

		comments[i].Story.LikesCount = uint(storyLikesCount)
		comments[i].Story.IsLikedByUser = isStoryLikedByUser
		comments[i].Story.CommentsCount = uint(storyCommentsCount)
		comments[i].Story.IsEditableByUser = uint(userId) == comments[i].Story.UserID
		comments[i].Story.IsDeletableByUser = uint(userId) == comments[i].Story.UserID
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

func GetUserPrivateStoriesCount(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	targetUserId := userId
	if userIdParam := c.Param("id"); userIdParam != "" {
		parsedId, err := strconv.ParseUint(userIdParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.UserNotFound, Data: nil})
			return
		}

		targetUserId = uint(parsedId)
	}

	if targetUserId != userId {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	var privateStoriesCount int64
	if err := database.DB.Model(&models.Story{}).Where("user_id = ? AND is_private = ?", userId, true).Count(&privateStoriesCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	maxPrivateStories := 3
	if user.IsPremium {
		maxPrivateStories = -1 // -1 means unlimited
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: map[string]any{
		"count": privateStoriesCount,
		"max":   maxPrivateStories,
	}})
}
