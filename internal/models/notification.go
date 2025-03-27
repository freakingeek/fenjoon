package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"-"`
	Title     string         `gorm:"default ''" json:"title"`
	Message   string         `gorm:"default ''" json:"message"`
	IsRead    bool           `gorm:"default false" json:"isRead"`
	Image     string         `gorm:"default ''" json:"image"`
	Url       string         `gorm:"default ''" json:"url"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
