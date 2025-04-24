package main

import (
	"log"
	"os"

	"github.com/freakingeek/fenjoon/internal/database"
	"github.com/freakingeek/fenjoon/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	database.InitDB()
	database.InitRedis()

	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://app.fnjo.ir", "https://app.stage.fnjo.ir", "https://fenjoon.vercel.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	err := r.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
