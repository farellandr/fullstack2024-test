package main

import (
	"log"

	"github.com/farellandr/fullstack2024-test/config"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	_, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router := gin.Default()

	router.Run(":3222")
}
