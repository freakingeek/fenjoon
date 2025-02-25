package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Phone      string         `gorm:"varchar(11);<-:create" json:"phone"`
	FirstName  string         `gorm:"varchar(50);not null" json:"firstName"`
	LastName   string         `gorm:"varchar(50);not null" json:"lastName"`
	Nickname   string         `gorm:"varchar(50);not null" json:"nickname"`
	Stories    []Story        `gorm:"foreignKey:UserID" json:"-"`
	IsVerified bool           `gorm:"default false" json:"isVerified"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}
