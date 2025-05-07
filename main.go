package main

import (
	"log"

	"github.com/farellandr/fullstack2024-test/config"
	"github.com/farellandr/fullstack2024-test/handlers"
	"github.com/farellandr/fullstack2024-test/models"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&models.Client{})

	router := gin.Default()

	clientHandler := handlers.NewClientHandler(db)
	api := router.Group("/api/v1")
	{
		api.POST("/clients", clientHandler.CreateClient)
		api.GET("/clients", clientHandler.GetAllClients)
		api.GET("/clients/:slug", clientHandler.GetClientBySlug)
		api.PUT("/clients/:slug", clientHandler.UpdateClient)
		api.DELETE("/clients/:slug", clientHandler.DeleteClient)
	}

	router.Run(":3222")
}
