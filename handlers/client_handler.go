package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/farellandr/fullstack2024-test/models"
	"github.com/farellandr/fullstack2024-test/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientHandler struct {
	DB          *gorm.DB
	RedisClient *utils.RedisClient
	S3Service   *utils.S3Service
}

func NewClientHandler(db *gorm.DB, redisClient *utils.RedisClient, s3Service *utils.S3Service) *ClientHandler {
	return &ClientHandler{
		DB:          db,
		RedisClient: redisClient,
		S3Service:   s3Service,
	}
}

// CreateClient godoc
// @Summary Create a new client
// @Description Creates a new client record in the database and caches it in Redis.
// @Tags clients
// @Accept json
// @Produce json
// @Param client body models.Client true "Client object to be created"
// @Success 201 {object} models.Client "Successfully created client"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 500 {object} map[string]string "Failed to create client"
// @Router /clients [post]
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

// GetAllClients godoc
// @Summary Get all clients
// @Description Retrieves a list of all client records from the database.
// @Tags clients
// @Produce json
// @Success 200 {array} models.Client "A list of all clients"
// @Failure 500 {object} map[string]string "Failed to retrieve clients"
// @Router /clients [get]
func (h *ClientHandler) GetAllClients(c *gin.Context) {
	var clients []models.Client

	if err := h.DB.Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clients"})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// GetClientBySlug godoc
// @Summary Get client by slug
// @Description Retrieves a single client record by its unique slug, first checking the Redis cache and then the database.
// @Tags clients
// @Produce json
// @Param slug path string true "The unique slug of the client to retrieve"
// @Success 200 {object} models.Client "The requested client"
// @Failure 404 {object} map[string]string "Client not found"
// @Router /clients/{slug} [get]
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

// UpdateClient godoc
// @Summary Update a client
// @Description Updates an existing client's data identified by its slug and refreshes the Redis cache.
// @Tags clients
// @Accept json
// @Produce json
// @Param slug path string true "The unique slug of the client to update"
// @Param client body models.Client true "Updated client object"
// @Success 200 {object} models.Client "Successfully updated client"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Failed to update client"
// @Router /clients/{slug} [put]
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

// DeleteClient godoc
// @Summary Delete a client
// @Description Deletes a client record from the database and removes it from the Redis cache by its unique slug.
// @Tags clients
// @Produce json
// @Param slug path string true "The unique slug of the client to delete"
// @Success 200 {object} map[string]string "Successfully deleted client"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Failed to delete client"
// @Router /clients/{slug} [delete]
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

// UploadClientLogo godoc
// @Summary Upload client logo
// @Description Uploads a client logo image to S3, updates the client record in the database with the S3 URL, and refreshes the Redis cache.
// @Tags clients
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "The unique slug of the client to update the logo for"
// @Param logo formData file true "The logo file to upload (e.g., .png, .jpg)"
// @Success 200 {object} map[string]string "Successfully uploaded logo"
// @Failure 400 {object} map[string]string "Invalid file upload"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Failed to upload logo or update client"
// @Router /clients/{slug}/upload-logo [post]
func (h *ClientHandler) UploadClientLogo(c *gin.Context) {
	slug := c.Param("slug")
	var client models.Client

	if err := h.DB.Where("slug = ?", slug).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	file, err := c.FormFile("logo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	if h.S3Service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S3 service not available"})
		return
	}

	logoURL, err := h.S3Service.UploadFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	if err := h.DB.Model(&client).Update("client_logo", logoURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update client logo"})
		return
	}

	if h.RedisClient != nil {
		if err := h.DB.Where("slug = ?", slug).First(&client).Error; err != nil {
			log.Printf("Warning: Failed to get updated client for Redis: %v", err)
		} else {
			if err := h.RedisClient.SetClientData(slug, client); err != nil {
				log.Printf("Warning: Failed to update client in Redis: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Logo uploaded successfully",
		"logo_url": logoURL,
	})
}
