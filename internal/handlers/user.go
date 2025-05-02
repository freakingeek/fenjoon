package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/freakingeek/fenjoon/internal/utils"
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
			"id":         user.ID,
			"firstName":  user.FirstName,
			"lastName":   user.LastName,
			"nickname":   user.Nickname,
			"phone":      user.Phone,
			"isBot":      user.IsBot,
			"isVerified": user.IsVerified,
			"isPremium":  user.IsPremium,
			"bio":        user.Bio,
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

func GetCurrentUserStories(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
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

	query := database.DB.Model(&models.Story{}).Where("user_id = ?", userId)

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
		stories[i].IsPrivatableByUser = userId == stories[i].UserID
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

func GetUserById(c *gin.Context) {
	userId, _ := auth.GetUserIdFromContext(c)

	targetUserIdUint64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralBadRequest,
			Data:    nil,
		})
		return
	}
	targetUserId := uint(targetUserIdUint64)

	var user models.User
	if err := database.DB.First(&user, targetUserId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	var followersCount int64
	if err := database.DB.Model(&models.Follow{}).
		Where("following_id = ?", targetUserId).
		Where("deleted_at IS NULL").
		Count(&followersCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var followingsCount int64
	if err := database.DB.Model(&models.Follow{}).
		Where("follower_id = ?", targetUserId).
		Where("deleted_at IS NULL").
		Count(&followingsCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	isFollowedByUser := false
	if userId != 0 {
		var count int64
		if err := database.DB.Model(&models.Follow{}).
			Where("follower_id = ? AND following_id = ?", userId, targetUserId).
			Where("deleted_at IS NULL").
			Count(&count).Error; err == nil && count > 0 {
			isFollowedByUser = true
		}
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"id":               user.ID,
			"firstName":        user.FirstName,
			"lastName":         user.LastName,
			"nickname":         user.Nickname,
			"bio":              user.Bio,
			"followersCount":   followersCount,
			"followingsCount":  followingsCount,
			"isFollowedByUser": isFollowedByUser,
		},
	})
}

func GetUserPublicStories(c *gin.Context) {
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

	query := database.DB.Model(&models.Story{}).Where("user_id = ? AND is_private = ?", targetUserId, false)

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
		stories[i].IsPrivatableByUser = userId == stories[i].UserID
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

func FollowUser(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	followingUserId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	if userId == uint(followingUserId) {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.UserFollowSelf, Data: nil})
		return
	}

	var existingFollow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", userId, followingUserId).First(&existingFollow).Error; err == nil {
		c.JSON(http.StatusConflict, responses.ApiResponse{Status: http.StatusConflict, Message: messages.UserAlreadyFollowed, Data: nil})
		return
	}

	follow := models.Follow{FollowerID: uint(userId), FollowingID: uint(followingUserId)}
	if err := database.DB.Create(&follow).First(&follow, follow.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err == nil {
		text := fmt.Sprintf("%s از حالا دنبالت میکنه!", utils.GetUserDisplayName(user))

		notification := models.Notification{UserID: uint(followingUserId), Title: "دنبال کننده جدید داری!", Message: text, Url: fmt.Sprintf("/author/%d", userId)}
		if err := services.SendInAppNotification(notification); err != nil {
			fmt.Printf("Failed to send in-app notification: %v\n", err)
		}

		var pushToken models.PushToken
		if err := database.DB.Where("user_id = ?", followingUserId).First(&pushToken).Error; err == nil {
			if err := services.SendPushNotification([]string{pushToken.Token}, text); err != nil {
				fmt.Printf("Failed to send push notification: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: true})
}

func UnfollowUser(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	followingUserId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.UserNotFound, Data: nil})
		return
	}

	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", userId, followingUserId).First(&follow).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.GeneralNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: true})
}

func GetUserFollowers(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64
	if err := database.DB.
		Model(&models.Follow{}).
		Where("following_id = ?", userId).
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var followers []models.User
	if err := database.DB.
		Table("users").
		Select("users.*").
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.following_id = ?", userId).
		Where("follows.deleted_at IS NULL").
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&followers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"followers": followers,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}

func GetUserFollowings(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int64
	if err := database.DB.
		Model(&models.Follow{}).
		Where("follower_id = ?", userId).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var followings []models.User
	if err := database.DB.
		Table("users").
		Select("users.*").
		Joins("JOIN follows ON follows.following_id = users.id").
		Where("follows.follower_id = ?", userId).
		Where("follows.deleted_at IS NULL").
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&followings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"followings": followings,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}
