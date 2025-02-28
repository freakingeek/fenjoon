package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	StoryId   uint           `gorm:"not null" json:"-"`
	UserId    uint           `gorm:"not null" json:"-"`
	Text      string         `gorm:"no null" json:"text"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
