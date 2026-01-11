// Package handlers contains HTTP request handlers.
package handlers

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vknow360/otaship/backend/internal/database"
)

// AssetsHandler handles the /api/assets endpoint.
// Serves static asset files for update bundles.
type AssetsHandler struct {
	updatesDir string
	updateRepo *database.UpdateRepository
}

// NewAssetsHandler creates a new assets handler.
func NewAssetsHandler(updatesDir string, updateRepo *database.UpdateRepository) *AssetsHandler {
	return &AssetsHandler{
		updatesDir: updatesDir,
		updateRepo: updateRepo,
	}
}

// Handle serves asset files to Expo clients.
func (h *AssetsHandler) Handle(c *gin.Context) {
	assetName := c.Query("asset")
	platform := c.Query("platform")
	runtimeVersion := c.Query("runtimeVersion")
	redirectUrl := c.Query("redirect")
	updateId := c.Query("updateId")
	isLaunchAsset := c.Query("isLaunchAsset")

	// Redirect logic for Cloudinary / Analytics
	if redirectUrl != "" {
		// If it's the launch asset (bundle), track it
		if isLaunchAsset == "true" && updateId != "" && h.updateRepo != nil {
			// Increment in background to not block redirect
			go func() {
				_ = h.updateRepo.IncrementDownloads(context.Background(), updateId)
			}()
		}

		c.Redirect(http.StatusFound, redirectUrl)
		return
	}

	// Validate parameters (Legacy / Local Fallback)
	if assetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No asset name provided."})
		return
	}

	if platform != "ios" && platform != "android" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No platform provided. Expected \"ios\" or \"android\".",
		})
		return
	}

	if runtimeVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No runtimeVersion provided."})
		return
	}

	// Security: Ensure asset path doesn't escape updates directory
	cleanPath := filepath.Clean(assetName)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset path."})
		return
	}

	// Resolve full path by joining with updates directory
	assetPath := filepath.Join(h.updatesDir, filepath.FromSlash(cleanPath))

	// Check file exists
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Asset \"" + assetName + "\" does not exist.",
		})
		return
	}

	// Open file
	file, err := os.Open(assetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open asset."})
		return
	}
	defer file.Close()

	// Determine content type
	contentType := h.getContentType(assetPath)

	// Get file size
	stat, _ := file.Stat()
	fileSize := stat.Size()

	// Set headers and stream file
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.DataFromReader(http.StatusOK, fileSize, contentType, file, nil)
}

// getContentType determines the MIME type based on file extension.
func (h *AssetsHandler) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	// JavaScript bundles
	case ".js", ".hbc":
		return "application/javascript"
	case ".bundle":
		return "application/javascript"

	// Images
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"

	// Fonts
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"

	// JSON
	case ".json":
		return "application/json"

	// Default
	default:
		return "application/octet-stream"
	}
}
