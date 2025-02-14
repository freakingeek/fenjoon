package handlers

import (
	"net/http"
	"regexp"
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
	re := regexp.MustCompile(`^[\p{Arabic}\s]+$`)
	return re.MatchString(text)
}

func GetUserById(c *gin.Context) {
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

	var user models.User

	if err := database.DB.Where("id = ?", claims["id"]).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{
			Status:  http.StatusNotFound,
			Message: messages.UserNotFound,
			Data:    map[string]interface{}{"user": nil},
		})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data:    map[string]interface{}{"user": user},
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
	if strings.TrimSpace(request.FirstName) != "" {
		updates["first_name"] = request.FirstName
	}
	if strings.TrimSpace(request.LastName) != "" {
		updates["last_name"] = request.LastName
	}
	if strings.TrimSpace(request.Nickname) != "" {
		updates["nickname"] = request.Nickname
	}

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
		Data:    map[string]interface{}{"user": user},
	})
}
