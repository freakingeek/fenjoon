package services

import (
	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/models"
)

func SendInAppNotification(notification models.Notification) error {
	if err := database.DB.Create(&notification).Error; err != nil {
		return err
	}

	return nil
}
