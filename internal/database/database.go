package database

import (
	"log"
	"os"

	"github.com/freakingeek/fenjoon/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Story{}, &models.Like{}, &models.Comment{}, &models.Share{}, &models.PushToken{}, &models.CommentLike{}, &models.Notification{}, &models.StoryReport{}, &models.Follow{})
	if err != nil {
		log.Fatal("failed to migrate database", err)
	}

	DB = db
	log.Println("Database connected!")
}
