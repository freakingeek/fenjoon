package models

import (
	"time"

	"gorm.io/gorm"
)

type PushToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"userId"`
	Token     string         `gorm:"not null" json:"token"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
