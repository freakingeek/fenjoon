package models

import (
	"time"

	"gorm.io/gorm"
)

type Like struct {
	ID        uint           `gorm:"primaryKey" json:"-"`
	StoryID   uint           `gorm:"not null" json:"-"`
	UserID    uint           `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
