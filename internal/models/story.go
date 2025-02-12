package models

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Text      string `gorm:"type:varchar(256); not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
