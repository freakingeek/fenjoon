package models

import (
	"time"

	"gorm.io/gorm"
)

type Bookmark struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	StoryID   uint           `gorm:"not null" json:"-"`
	UserID    uint           `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
