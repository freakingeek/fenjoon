package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/freakingeek/fenjoon/internal/auth"
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/messages"
	"github.com/freakingeek/fenjoon/internal/models"
	"github.com/freakingeek/fenjoon/internal/responses"
	"github.com/gin-gonic/gin"
)

func GetStoryReports(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	var reports []models.StoryReport
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

	if err := database.DB.Model(&models.StoryReport{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Preload("Story.User").Order("id DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{
		Status:  http.StatusOK,
		Message: messages.GeneralSuccess,
		Data: map[string]any{
			"reports": reports,
			"pagination": map[string]any{
				"total": total,
				"page":  page,
				"limit": limit,
				"pages": int((total + int64(limit) - 1) / int64(limit)),
			},
		},
	})
}

func GetStoryReport(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	reportId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var report models.StoryReport
	if err := database.DB.Preload("Story.User").Where("id = ?", reportId).First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.ReportNotFound, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: report})
}

func ResolveStoryReport(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	var request struct {
		Reason string `json:"reason" binding:"required,min=5,max=250"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	reportId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var report models.StoryReport
	if err := database.DB.Where("id = ?", reportId).First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.ReportNotFound, Data: nil})
		return
	}

	report.ResolvedAt = time.Now()
	report.ResolvedBy = userId
	report.ResolutionNotes = request.Reason
	report.Status = "resolved"

	if err := database.DB.Save(report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var story models.Story
	if err := database.DB.First(&story, report.StoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Delete(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: story})
}

func RejectStoryReport(c *gin.Context) {
	userId, err := auth.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ApiResponse{Status: http.StatusUnauthorized, Message: messages.GeneralUnauthorized, Data: nil})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, responses.ApiResponse{Status: http.StatusForbidden, Message: messages.GeneralAccessDenied, Data: nil})
		return
	}

	var request struct {
		Reason string `json:"reason" binding:"required,min=5,max=250"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.GeneralBadRequest, Data: nil})
		return
	}

	reportId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ApiResponse{Status: http.StatusBadRequest, Message: messages.StoryNotFound, Data: nil})
		return
	}

	var report models.StoryReport
	if err := database.DB.Where("id = ?", reportId).First(&report).Error; err != nil {
		c.JSON(http.StatusNotFound, responses.ApiResponse{Status: http.StatusNotFound, Message: messages.ReportNotFound, Data: nil})
		return
	}

	report.ResolvedAt = time.Now()
	report.ResolvedBy = userId
	report.ResolutionNotes = request.Reason
	report.Status = "rejected"

	if err := database.DB.Save(report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	var story models.Story
	if err := database.DB.Unscoped().First(&story, report.StoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	if err := database.DB.Unscoped().Model(&story).Update("deleted_at", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, responses.ApiResponse{Status: http.StatusInternalServerError, Message: messages.GeneralFailed, Data: nil})
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponse{Status: http.StatusOK, Message: messages.GeneralSuccess, Data: story})
}
