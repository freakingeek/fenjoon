package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Phone            string         `gorm:"varchar(11);<-:create" json:"-"`
	Bio              string         `gorm:"varchar(100)" json:"bio"`
	FollowersCount   uint           `gorm:"-" json:"followersCount"`
	FollowingsCount  uint           `gorm:"-" json:"followingsCount"`
	IsFollowedByUser bool           `gorm:"default:false" json:"isFollowedByUser"`
	FirstName        string         `gorm:"varchar(50);not null" json:"firstName"`
	LastName         string         `gorm:"varchar(50);not null" json:"lastName"`
	Nickname         string         `gorm:"varchar(50);not null" json:"nickname"`
	Stories          []Story        `gorm:"foreignKey:UserID" json:"-"`
	Notifications    []Notification `gorm:"foreignKey:UserID" json:"-"`
	IsVerified       bool           `gorm:"default:false" json:"isVerified"`
	IsBot            bool           `gorm:"default:false" json:"isBot"`
	IsAdmin          bool           `gorm:"default:false" json:"-"`
	IsPremium        bool           `gorm:"default:false" json:"isPremium"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
