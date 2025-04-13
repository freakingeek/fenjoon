package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	StoryID           uint           `gorm:"not null" json:"-"`
	UserID            uint           `gorm:"not null" json:"-"`
	User              User           `gorm:"foreignKey:UserID" json:"user"`
	Story             Story          `gorm:"foreignKey:StoryID" json:"story"`
	Text              string         `gorm:"type:varchar(500);not null" json:"text"`
	Likes             []CommentLike  `gorm:"foreignKey:CommentID" json:"-"`
	LikesCount        uint           `gorm:"-" json:"likesCount"`
	IsLikedByUser     bool           `gorm:"-" json:"isLikedByUser"`
	IsEditableByUser  bool           `gorm:"-" json:"isEditableByUser"`
	IsDeletableByUser bool           `gorm:"-" json:"isDeletableByUser"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"-"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}
