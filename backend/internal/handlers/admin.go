// Package handlers contains HTTP request handlers.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/models"
	"github.com/vknow360/otaship/backend/internal/services"
	"github.com/vknow360/otaship/backend/internal/storage"
	"github.com/vknow360/otaship/backend/internal/utils"
)

// AdminHandler handles admin API endpoints.
type AdminHandler struct {
	updateRepo        *database.UpdateRepository
	analyticsRepo     *database.AnalyticsRepository
	cloudinaryService *storage.CloudinaryService
	projectRepo       *database.ProjectRepository
	apiKeyRepo        *database.APIKeyRepository
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(
	updateRepo *database.UpdateRepository,
	analyticsRepo *database.AnalyticsRepository,
	cloudinaryService *storage.CloudinaryService,
	projectRepo *database.ProjectRepository,
	apiKeyRepo *database.APIKeyRepository,
) *AdminHandler {
	return &AdminHandler{
		updateRepo:        updateRepo,
		analyticsRepo:     analyticsRepo,
		cloudinaryService: cloudinaryService,
		projectRepo:       projectRepo,
		apiKeyRepo:        apiKeyRepo,
	}
}

// ListUpdates returns a list of all registered updates.
// GET /api/admin/updates
func (h *AdminHandler) ListUpdates(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse query params
	limit := int64(50)
	offset := int64(0)

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Build filter
	filter := bson.M{}
	if channel := c.Query("channel"); channel != "" {
		filter["channel"] = channel
	}
	if runtime := c.Query("runtimeVersion"); runtime != "" {
		filter["runtimeVersion"] = runtime
	}

	// Check if database is connected
	if h.updateRepo == nil {
		c.JSON(http.StatusOK, gin.H{
			"updates": []interface{}{},
			"total":   0,
			"message": "Database not connected. Updates are served from filesystem.",
		})
		return
	}

	updates, total, err := h.updateRepo.FindAll(ctx, filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updates": updates,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// RegisterUpdate registers a new update in the database.
// POST /api/admin/updates
func (h *AdminHandler) RegisterUpdate(c *gin.Context) {
	if h.updateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database not connected",
		})
		return
	}

	// 1. Parse Multipart Form
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { // 100 MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form: " + err.Error()})
		return
	}

	projectSlug := c.PostForm("projectSlug")
	runtimeVersion := c.PostForm("runtimeVersion")
	channel := c.DefaultPostForm("channel", models.ChannelProduction)
	platform := c.DefaultPostForm("platform", models.PlatformAll)
	rolloutStr := c.DefaultPostForm("rolloutPercentage", "100")
	updateID := c.PostForm("updateId")

	if projectSlug == "" || runtimeVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "projectSlug and runtimeVersion are required"})
		return
	}

	rolloutPercentage, _ := strconv.Atoi(rolloutStr)

	// 2. Handle File Upload
	file, header, err := c.Request.FormFile("bundle")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bundle file is required: " + err.Error()})
		return
	}
	defer file.Close()

	if updateID == "" {
		updateID = services.GenerateUUID()
	}

	// Create temp dir for this update
	tempDir, err := os.MkdirTemp("", "otaship-upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp dir"})
		return
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// Save zip file
	zipPath := filepath.Join(tempDir, header.Filename)
	out, err := os.Create(zipPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save zip file"})
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write zip file"})
		return
	}
	out.Close() // Close explicitly before unzip

	// Unzip
	bundlePath := filepath.Join(tempDir, "extracted")
	if err := utils.Unzip(zipPath, bundlePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unzip bundle: " + err.Error()})
		return
	}

	// Look for unnecessary nesting (common when zipping folders, especially from CLI)
	// 1. If entries contains only ONE folder, use that as root
	entries, err := os.ReadDir(bundlePath)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		bundlePath = filepath.Join(bundlePath, entries[0].Name())
	} else {
		// 2. Fallback: If "dist" folder exists and "metadata.json" is NOT in root, try "dist"
		if _, err := os.Stat(filepath.Join(bundlePath, "metadata.json")); os.IsNotExist(err) {
			distPath := filepath.Join(bundlePath, "dist")
			if info, err := os.Stat(distPath); err == nil && info.IsDir() {
				// Check if metadata exists inside dist
				if _, err := os.Stat(filepath.Join(distPath, "metadata.json")); err == nil {
					bundlePath = distPath
				}
			}
		}
	}

	update := &models.Update{
		ProjectSlug:       projectSlug,
		UpdateID:          updateID,
		RuntimeVersion:    runtimeVersion,
		Channel:           channel,
		Platform:          platform,
		BundlePath:        bundlePath, // Pointing to temp path for processing
		RolloutPercentage: rolloutPercentage,
		IsActive:          true,
		IsRollback:        false,
	}

	// Auto-create project if it doesn't exist
	if h.projectRepo != nil {
		h.projectRepo.EnsureProjectExists(projectSlug, projectSlug)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // Increased timeout for upload
	defer cancel()

	// ---------------------------------------------------------
	// NEW LOGIC: Parse Metadata ALWAYS (before Cloudinary)
	// ---------------------------------------------------------
	metadataPath := filepath.Join(bundlePath, "metadata.json")
	metadataFile, err := os.ReadFile(metadataPath)
	if err == nil {
		var metadata models.UpdateMetadata
		if err := json.Unmarshal(metadataFile, &metadata); err == nil {
			// Read expoConfig.json
			expoConfigPath := filepath.Join(bundlePath, "expoConfig.json")
			if expoConfigFile, err := os.ReadFile(expoConfigPath); err == nil {
				var expoConfig map[string]interface{}
				if err := json.Unmarshal(expoConfigFile, &expoConfig); err == nil {
					metadata.ExpoConfig = expoConfig
				}
			}
			update.Metadata = &metadata

			// Compute hashes (Initial pass locally)
			for platform, pm := range metadata.FileMetadata {
				// Bundle Hash
				bundlePathFull := filepath.Join(bundlePath, pm.Bundle)
				if data, err := os.ReadFile(bundlePathFull); err == nil {
					pm.BundleKey = services.ComputeSHA256Hash(data)[:32]
					pm.BundleHash = services.Base64URLEncode(services.ComputeSHA256HashBytes(data))
				}

				// Assets Hashes
				for i, asset := range pm.Assets {
					assetPath := filepath.Join(bundlePath, asset.Path)
					if data, err := os.ReadFile(assetPath); err == nil {
						pm.Assets[i].Key = services.ComputeSHA256Hash(data)[:32]
						pm.Assets[i].Hash = services.Base64URLEncode(services.ComputeSHA256HashBytes(data))
					}
				}
				metadata.FileMetadata[platform] = pm
			}
			update.Metadata = &metadata
		}
	} else {
		log.Printf("Warning: Failed to read metadata.json: %v", err)
	}

	// Upload to Cloudinary if connected
	if h.cloudinaryService != nil && h.cloudinaryService.IsConnected() {
		// Upload directory
		cloudFolder := fmt.Sprintf("updates/%s/%s", runtimeVersion, updateID)
		urlMap, err := h.cloudinaryService.UploadDirectory(ctx, cloudFolder, bundlePath)
		if err == nil {
			if update.Metadata != nil {
				metadata := *update.Metadata
				// Inject Cloudinary URLs if available
				if urlMap != nil {
					for platform, pm := range metadata.FileMetadata {
						// Update BundleUrl
						if url, ok := urlMap[filepath.FromSlash(pm.Bundle)]; ok {
							pm.BundleUrl = url
						}

						// Update Assets Urls
						for i, asset := range pm.Assets {
							if url, ok := urlMap[filepath.FromSlash(asset.Path)]; ok {
								pm.Assets[i].Url = url
							}
						}
						metadata.FileMetadata[platform] = pm
					}
				}
				update.Metadata = &metadata
			}
		} else {
			// Log error but don't fail, maybe we can run without cloudinary?
			// But if we are remote, we kinda need cloudinary or persistent storage.
			// For now, assuming Cloudinary is REQUIRED for remote setup.
			fmt.Printf("Error uploading to Cloudinary: %v\n", err)
		}
	} else {
		// Warning: Local storage on ephemeral filesystem (like Render) will be lost!
		log.Println("WARNING: Cloudinary not connected. Files uploaded to temp storage will be lost on restart.")
	}

	// Important: We cannot rely on BundlePath being valid after this request ends
	// if we are using ephemeral storage and not persisting it.
	// However, the ManifestHandler might need it if not using Cloudinary strings.
	// Since we are fixing for Render -> User should have Cloudinary or S3.
	// We will proceed assuming Cloudinary took care of file persistence via URLs.

	// Nulify BundlePath to avoid trying to read from it later if it's deleted?
	// Or keep it for the duration of the request?
	// Ideally update.BundlePath should be empty if we rely purely on Cloudinary URLs.

	if err := h.updateRepo.Create(ctx, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Update registered successfully",
		"update":  update,
	})
}

// UpdateUpdate modifies an existing update.
// PATCH /api/admin/updates/:id
func (h *AdminHandler) UpdateUpdate(c *gin.Context) {
	if h.updateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not connected"})
		return
	}

	id := c.Param("id")

	var req struct {
		IsActive          *bool `json:"isActive"`
		RolloutPercentage *int  `json:"rolloutPercentage"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updates := bson.M{}
	if req.IsActive != nil {
		updates["isActive"] = *req.IsActive
	}
	if req.RolloutPercentage != nil {
		updates["rolloutPercentage"] = *req.RolloutPercentage
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updates provided"})
		return
	}

	if err := h.updateRepo.Update(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Update modified successfully",
		"id":      id,
	})
}

// CreateRollback creates a rollback directive for an update.
// POST /api/admin/updates/:id/rollback
func (h *AdminHandler) CreateRollback(c *gin.Context) {
	if h.updateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not connected"})
		return
	}

	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a rollback update
	rollback := &models.Update{
		RuntimeVersion:    c.Query("runtimeVersion"),
		Channel:           c.DefaultQuery("channel", models.ChannelProduction),
		Platform:          models.PlatformAll,
		IsActive:          true,
		IsRollback:        true,
		RolloutPercentage: 100,
	}

	if err := h.updateRepo.Create(ctx, rollback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Rollback created successfully",
		"previousId": id,
		"rollback":   rollback,
	})
}

// GetStats returns analytics data.
// GET /api/admin/stats
func (h *AdminHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if h.analyticsRepo == nil {
		c.JSON(http.StatusOK, gin.H{
			"totalDownloads": 0,
			"todayDownloads": 0,
			"weekDownloads":  0,
			"byPlatform":     map[string]int{"android": 0, "ios": 0},
			"byChannel":      map[string]int{"production": 0, "staging": 0, "beta": 0},
			"message":        "Database not connected.",
		})
		return
	}

	summary, err := h.analyticsRepo.GetSummary(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// DeleteUpdate deletes an update and cleans up Cloudinary assets.
// DELETE /api/admin/updates/:id
func (h *AdminHandler) DeleteUpdate(c *gin.Context) {
	if h.updateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not connected"})
		return
	}

	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch update before deleting for Cloudinary cleanup
	update, err := h.updateRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Update not found"})
		return
	}

	// 2. Delete from DB
	if err := h.updateRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Cleanup Cloudinary in background
	if update != nil && h.cloudinaryService.IsConnected() {
		go func() {
			folder := fmt.Sprintf("updates/%s/%s", update.RuntimeVersion, update.UpdateID)
			ctxBg := context.Background()
			if err := h.cloudinaryService.DeleteFolder(ctxBg, folder); err != nil {
				fmt.Printf("Warning: Failed to cleanup Cloudinary folder %s: %v\n", folder, err)
			}
		}()
	}

	c.Status(http.StatusOK)
}

// ListProjects returns all projects.
// GET /api/admin/projects
func (h *AdminHandler) ListProjects(c *gin.Context) {
	projects, err := h.projectRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"total":    len(projects),
	})
}

// CreateProject creates a new project.
// POST /api/admin/projects
func (h *AdminHandler) CreateProject(c *gin.Context) {
	var req struct {
		Slug        string `json:"slug" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &models.Project{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.projectRepo.Create(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// DeleteProject deletes a project and all its updates.
// DELETE /api/admin/projects/:slug
func (h *AdminHandler) DeleteProject(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	// Delete all updates for this project
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all updates for this project to cleanup Cloudinary
	if h.updateRepo != nil {
		updates, _, _ := h.updateRepo.FindAll(ctx, bson.M{"projectSlug": slug}, 1000, 0)
		for _, update := range updates {
			if h.cloudinaryService.IsConnected() {
				folder := fmt.Sprintf("updates/%s/%s", update.RuntimeVersion, update.UpdateID)
				go h.cloudinaryService.DeleteFolder(context.Background(), folder)
			}
		}
		// Delete updates from DB
		h.updateRepo.DeleteByProjectSlug(ctx, slug)
	}

	// Delete the project
	if err := h.projectRepo.Delete(slug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	keys, err := h.apiKeyRepo.FindAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":  keys,
		"total": len(keys),
	})
}

func (h *AdminHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name   string   `json:"name" binding:"required"`
		Scopes []string `json:"scopes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plainKey, apiKey, err := h.apiKeyRepo.Create(context.Background(), req.Name, req.Scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":    plainKey,
		"apiKey": apiKey,
	})
}

func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	if err := h.apiKeyRepo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
