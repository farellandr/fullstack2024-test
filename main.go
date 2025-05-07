package main

import (
	"log"

	"github.com/farellandr/fullstack2024-test/config"
	_ "github.com/farellandr/fullstack2024-test/docs"
	"github.com/farellandr/fullstack2024-test/handlers"
	"github.com/farellandr/fullstack2024-test/models"
	"github.com/farellandr/fullstack2024-test/utils"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/joho/godotenv"
)

// @title Fullstack2024 Test API
// @version 1.0
// @description ASI Asia Pacific Fullstack test.
// @host localhost:3222
// @BasePath /api/v1
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&models.Client{})

	redisClient := utils.InitRedis()
	s3Service := utils.InitS3()

	router := gin.Default()

	clientHandler := handlers.NewClientHandler(db, redisClient, s3Service)
	api := router.Group("/api/v1")
	{
		api.POST("/clients", clientHandler.CreateClient)
		api.GET("/clients", clientHandler.GetAllClients)
		api.GET("/clients/:slug", clientHandler.GetClientBySlug)
		api.PUT("/clients/:slug", clientHandler.UpdateClient)
		api.DELETE("/clients/:slug", clientHandler.DeleteClient)
		api.POST("/clients/:slug/logo", clientHandler.UploadClientLogo)
	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":3222")
}
