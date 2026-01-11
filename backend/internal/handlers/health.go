// Package handlers contains HTTP request handlers.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/storage"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// HealthHandler handles GET /api/health requests.
type HealthHandler struct {
	version string
	db      *database.MongoDB
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(version string, db *database.MongoDB) *HealthHandler {
	return &HealthHandler{
		version: version,
		db:      db,
	}
}

// Handle returns the current health status of the server.
func (h *HealthHandler) Handle(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Version:   h.version,
		Timestamp: c.GetHeader("Date"),
		Services:  make(map[string]string),
	}

	// Check database
	if h.db != nil && h.db.IsConnected() {
		if err := h.db.HealthCheck(); err != nil {
			response.Services["database"] = "error: " + err.Error()
		} else {
			response.Services["database"] = "ok"
		}
	} else {
		response.Services["database"] = "not configured"
	}

	// Check Cloudinary
	if storage.Cloudinary != nil && storage.Cloudinary.IsConnected() {
		response.Services["cloudinary"] = "ok"
	} else {
		response.Services["cloudinary"] = "not configured"
	}

	response.Services["signing"] = "ok"

	c.JSON(http.StatusOK, response)
}
