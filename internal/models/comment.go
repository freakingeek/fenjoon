package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	StoryID   uint           `gorm:"not null" json:"-"`
	UserID    uint           `gorm:"not null" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Text      string         `gorm:"type:varchar(500);not null" json:"text"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
