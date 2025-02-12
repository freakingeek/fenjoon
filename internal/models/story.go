package models

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Text      string `gorm:"type:varchar(256); not null" json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}
