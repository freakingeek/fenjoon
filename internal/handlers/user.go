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
	// Regular expression to match only Persian characters
	re := regexp.MustCompile(`^[\p{Arabic}\s\x{200C}\x{0640}]+$`)
	return re.MatchString(text)
}

func GetUserById(c *gin.Context) {
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

	var user models.User

	if err := database.DB.Where("id = ?", claims["id"]).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.UserNotFound,
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]interface{}{
			"id":        user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"nickname":  user.Nickname,
			"phone":     user.Phone,
		},
	})
}

func GetUserStories(c *gin.Context) {
	token, err := auth.ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	claims, err := auth.ParseJWTToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	floatUserId, ok := claims["id"].(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	userId := uint(floatUserId)

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
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"stories": nil},
		})
		return
	}

	if err := database.DB.Where("user_id = ?", userId).Limit(limit).Offset(offset).Find(&stories).Error; err != nil {
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
				"pages": int((total + int64(limit) - 1) / int64(limit)), // Calculate total pages
			},
		},
	})
}

func UpdateUserById(c *gin.Context) {
	token, err := auth.ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	claims, err := auth.ParseJWTToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{
			Status:  http.StatusUnauthorized,
			Message: messages.GeneralUnauthorized,
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	var request struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Nickname  string `json:"nickname"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	if (request.FirstName != "" && !isFarsiText(request.FirstName)) ||
		(request.LastName != "" && !isFarsiText(request.LastName)) ||
		(request.Nickname != "" && !isFarsiText(request.Nickname)) {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{
			Status:  http.StatusBadRequest,
			Message: "نام، نام خانوادگی و نام مستعار باید فقط شامل حروف فارسی باشند.",
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	var user models.User

	if err := database.DB.Where("id = ?", claims["id"]).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.UserNotFound,
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	updates := map[string]interface{}{}
	updates["first_name"] = strings.TrimSpace(request.FirstName)
	updates["last_name"] = strings.TrimSpace(request.LastName)
	updates["nickname"] = strings.TrimSpace(request.Nickname)

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{
			Status:  http.StatusInternalServerError,
			Message: messages.GeneralFailed,
			Data:    map[string]interface{}{"user": nil},
		})
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.UserEdited,
		Data:    user,
	})
}
