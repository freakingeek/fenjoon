package models

import (
	"time"

	"gorm.io/gorm"
)

type Follow struct {
	ID          uint           `gorm:"primaryKey" json:"-"`
	FollowerID  uint           `gorm:"not null" json:"-"`
	FollowingID uint           `gorm:"not null" json:"-"`
	CreatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
