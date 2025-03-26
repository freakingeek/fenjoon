package models

import (
	"time"

	"gorm.io/gorm"
)

type CommentLike struct {
	ID        uint           `gorm:"primaryKey" json:"-"`
	CommentID uint           `gorm:"not null" json:"-"`
	UserID    uint           `gorm:"not null" json:"-"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
