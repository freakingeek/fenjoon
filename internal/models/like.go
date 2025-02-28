package models

import "time"

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	StoryId   uint      `gorm:"not null" json:"-"`
	UserId    uint      `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"-"`
}
