package utils

import (
	"fmt"

	"github.com/freakingeek/fenjoon/internal/models"
)

func GetUserDisplayName(user models.User) string {
	if user.Nickname != "" {
		return user.Nickname
	}

	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}

	if user.FirstName != "" || user.LastName != "" {
		return user.FirstName + user.LastName
	}

	return fmt.Sprintf("کاربر %d#", user.ID)
}
