package models

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Text       string         `gorm:"type:varchar(256);not null" json:"text"`
	UserID     uint           `gorm:"not null" json:"-"`
	User       User           `gorm:"foreignKey:UserID" json:"user"`
	Likes      []Like         `gorm:"foreignKey:StoryID" json:"-"`
	LikesCount uint           `gorm:"-" json:"likesCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
