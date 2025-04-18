package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/freakingeek/fenjoon/internal/services"
	"github.com/freakingeek/fenjoon/internal/utils"
	"github.com/gin-gonic/gin"
)

func GetCommentById(c *gin.Context) {
	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentNotFound, Data: nil})
		return
	}

	userId, _ := auth.GetUserIdFromContext(c)

	var comment models.Comment
	if err := database.DB.Preload("User").Where("id = ?", commentId).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var likesCount int64
	if err := database.DB.Model(&models.Like{}).Where("comment_id = ?", commentId).Count(&likesCount).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var isLikedByUser bool
	if err := database.DB.Model(&models.Like{}).
		Where("comment_id = ? AND user_id = ?", commentId, userId).
		Select("COUNT(*) > 0").
		Find(&isLikedByUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	comment.LikesCount = uint(likesCount)
	comment.IsLikedByUser = isLikedByUser
	comment.IsEditableByUser = userId == comment.UserID
	comment.IsDeletableByUser = userId == comment.UserID

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: comment})
}

func DeleteComment(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var comment models.Comment
	if err := database.DB.Preload("User").Where("id = ? AND user_id = ?", commentId, userId).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.CommentDeleted, Data: comment})
}

func UpdateComment(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var request struct {
		Text string `json:"text" binding:"required,min=5,max=250"`
	}

	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentNotFound, Data: nil})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentCharLimit, Data: nil})
		return
	}

	var comment models.Comment
	if err := database.DB.Preload("User").Where("id = ? AND user_id = ?", commentId, userId).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	comment.Text = request.Text

	if err := database.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.CommentEdited, Data: comment})
}

func LikeCommentById(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.UserNotFound, Data: nil})
		return
	}

	var existingLike models.CommentLike
	if err := database.DB.Where("comment_id = ? AND user_id = ?", commentId, userId).First(&existingLike).Error; err == nil {
		c.JSON(http.StatusConflict, responses.ApiResponse{Status: http.StatusConflict, Message: messages.CommentAlreadyLiked, Data: nil})
		return
	}

	commentLike := models.CommentLike{CommentID: uint(commentId), UserID: userId}
	if err := database.DB.Create(&commentLike).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if userId != comment.UserID {
		text := fmt.Sprintf("%s از نقدت خوشش اومد", utils.GetUserDisplayName(user))

		notification := models.Notification{UserID: comment.UserID, Title: "نقدت پسندیده شد!", Message: text, Url: fmt.Sprintf("/story/%d", comment.StoryID)}
		if err := services.SendInAppNotification(notification); err != nil {
			fmt.Printf("Failed to send in-app notification: %v\n", err)
		}

		var pushToken models.PushToken
		if err := database.DB.Where("user_id = ?", comment.UserID).First(&pushToken).Error; err == nil {
			if err := services.SendPushNotification([]string{pushToken.Token}, text); err != nil {
				fmt.Printf("Failed to send push notification: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.CommentLiked, Data: true})
}

func DislikeCommentById(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	var commentLike models.CommentLike
	if err := database.DB.Where("comment_id = ? AND user_id = ?", commentId, userId).First(&commentLike).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
		return
	}

	if err := database.DB.Delete(&commentLike).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.CommentDisliked, Data: true})
}

func GetCommentLikers(c *gin.Context) {
	commentId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.CommentNotFound, Data: nil})
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
		Table("comment_likes").
		Where("comment_id = ?", commentId).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var users []models.User
	if err := database.DB.
		Joins("JOIN comment_likes ON comment_likes.user_id = users.id").
		Where("comment_likes.comment_id = ?", commentId).
		Where("comment_likes.deleted_at IS NULL").
		Preload("Stories").
		Order("comment_likes.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"users": users,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}
