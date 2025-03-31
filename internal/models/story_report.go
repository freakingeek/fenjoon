package models

import (
	"time"

	"gorm.io/gorm"
)

type StoryReport struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	StoryID         uint           `gorm:"not null" json:"storyId"`
	UserID          uint           `gorm:"not null" json:"userId"`
	Reason          string         `gorm:"type:varchar(256);not null" json:"reason"`
	Status          string         `gorm:"type:varchar(64);not null;default:'pending'" json:"status"` // "pending", "resolved", "rejected"
	ResolvedAt      time.Time      `json:"resolvedAt,omitempty"`
	ResolvedBy      uint           `json:"resolvedBy,omitempty"`
	ResolutionNotes string         `gorm:"type:varchar(512);" json:"resolutionNotes,omitempty"`
	Story           Story          `gorm:"foreignKey:StoryID" json:"story,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
