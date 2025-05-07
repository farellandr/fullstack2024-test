package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/farellandr/fullstack2024-test/models"
	"github.com/farellandr/fullstack2024-test/utils"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClientHandler struct {
	DB          *gorm.DB
	RedisClient *utils.RedisClient
}

func NewClientHandler(db *gorm.DB, redisClient *utils.RedisClient) *ClientHandler {
	return &ClientHandler{
		DB:          db,
		RedisClient: redisClient,
	}
}

func (h *ClientHandler) CreateClient(c *gin.Context) {
	var client models.Client

	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if client.Slug == "" {
		client.Slug = generateSlug(client.Name)
	}

	if err := h.DB.Create(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client: " + err.Error()})
		return
	}

	if h.RedisClient != nil {
		if err := h.RedisClient.SetClientData(client.Slug, client); err != nil {
			log.Printf("Warning: Failed to save client to Redis: %v", err)
		}
	}

	c.JSON(http.StatusCreated, client)
}

func (h *ClientHandler) GetAllClients(c *gin.Context) {
	var clients []models.Client

	if err := h.DB.Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clients"})
		return
	}

	c.JSON(http.StatusOK, clients)
}

func (h *ClientHandler) GetClientBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var client models.Client

	if h.RedisClient != nil {
		data, err := h.RedisClient.GetClientData(slug)
		if err == nil {
			c.Data(http.StatusOK, "application/json", []byte(data))
			return
		}
	}

	if err := h.DB.Where("slug = ?", slug).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	if h.RedisClient != nil {
		if err := h.RedisClient.SetClientData(client.Slug, client); err != nil {
			log.Printf("Warning: Failed to save client to Redis: %v", err)
		}
	}

	c.JSON(http.StatusOK, client)
}

func (h *ClientHandler) UpdateClient(c *gin.Context) {
	slug := c.Param("slug")
	var client models.Client

	if err := h.DB.Where("slug = ?", slug).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	var updatedClient models.Client
	if err := c.ShouldBindJSON(&updatedClient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Model(&client).Updates(updatedClient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client"})
		return
	}

	if h.RedisClient != nil {
		if err := h.RedisClient.DeleteClientData(slug); err != nil {
			log.Printf("Warning: Failed to delete client from Redis: %v", err)
		}

		if client.Slug != slug {
			if err := h.RedisClient.SetClientData(client.Slug, client); err != nil {
				log.Printf("Warning: Failed to save client to Redis: %v", err)
			}
		} else {
			if err := h.RedisClient.SetClientData(slug, client); err != nil {
				log.Printf("Warning: Failed to save client to Redis: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, client)
}

func (h *ClientHandler) DeleteClient(c *gin.Context) {
	slug := c.Param("slug")
	var client models.Client

	if err := h.DB.Where("slug = ?", slug).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	if err := h.DB.Delete(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete client"})
		return
	}

	if h.RedisClient != nil {
		if err := h.RedisClient.DeleteClientData(slug); err != nil {
			log.Printf("Warning: Failed to delete client from Redis: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")

	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	shortUUID := uuid.New().String()[:8]
	return slug + "-" + shortUUID
}
