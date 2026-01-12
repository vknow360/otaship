// OTAShip - Production-ready OTA update server
//
// A high-performance OTA update server for Expo/React Native applications.
// Built with Go and Gin for reliability and scalability.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vknow360/otaship/backend/internal/config"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/handlers"
	"github.com/vknow360/otaship/backend/internal/middleware"
	"github.com/vknow360/otaship/backend/internal/services"
	"github.com/vknow360/otaship/backend/internal/storage"
)

// Version is the server version (set during build).
var Version = "1.0.0"

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Connect to MongoDB (optional - works without it)
	db, err := database.Connect(database.Config{
		URI:          cfg.MongoDBURI,
		DatabaseName: "otaship",
		Timeout:      10 * time.Second,
	})
	if err != nil {
		log.Printf("Warning: MongoDB connection failed: %v", err)
		log.Println("Running without database - admin features limited")
	}

	// Graceful shutdown
	if db != nil {
		defer db.Disconnect()
	}

	// Initialize Cloudinary (optional - works without it)
	cloudinaryService, err := storage.NewCloudinaryService(storage.Config{
		CloudName: cfg.CloudinaryCloudName,
		APIKey:    cfg.CloudinaryAPIKey,
		APISecret: cfg.CloudinaryAPISecret,
	})
	if err != nil {
		log.Printf("Warning: Cloudinary setup failed: %v", err)
	}
	_ = cloudinaryService // Will be used by handlers later

	// Determine updates directory
	updatesDir := getUpdatesDir()

	// Initialize repositories
	updateRepo := database.NewUpdateRepository(db)
	analyticsRepo := database.NewAnalyticsRepository(db)
	projectRepo := database.NewProjectRepository(db)
	apiKeyRepo := database.NewAPIKeyRepository(db)

	// Initialize services
	signingService := services.NewSigningService()
	if cfg.PrivateKeyPath != "" {
		if err := signingService.LoadPrivateKey(cfg.PrivateKeyPath); err != nil {
			log.Printf("Warning: Failed to load private key: %v", err)
			log.Printf("Code signing will be disabled")
		} else {
			log.Printf("Code signing enabled")
		}
	}

	rolloutService := services.NewRolloutService()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(Version, db)
	manifestHandler := handlers.NewManifestHandler(signingService, rolloutService, updateRepo, updatesDir)
	assetsHandler := handlers.NewAssetsHandler(updatesDir, updateRepo)
	adminHandler := handlers.NewAdminHandler(updateRepo, analyticsRepo, cloudinaryService, projectRepo, apiKeyRepo)

	// Create Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORS())

	// Health check (no auth)
	router.GET("/api/health", healthHandler.Handle)

	// Client endpoints (no auth - called by Expo apps)
	// NEW: Project-scoped manifest endpoint
	router.GET("/api/:projectSlug/manifest", manifestHandler.Handle)
	router.GET("/api/:projectSlug/assets", assetsHandler.Handle)
	// Legacy endpoints (for backward compatibility) - will return error without projectSlug
	router.GET("/api/manifest", manifestHandler.Handle)
	router.GET("/api/assets", assetsHandler.Handle)

	// Admin endpoints (protected)
	admin := router.Group("/api/admin")
	admin.Use(middleware.AdminAuth())
	{
		// Project management
		admin.GET("/projects", adminHandler.ListProjects)
		admin.POST("/projects", adminHandler.CreateProject)
		admin.DELETE("/projects/:slug", adminHandler.DeleteProject)

		// Updates (can filter by project query param)
		admin.GET("/updates", adminHandler.ListUpdates)
		admin.POST("/updates", adminHandler.RegisterUpdate)
		admin.POST("/assets/check", adminHandler.CheckAssets)
		admin.PATCH("/updates/:id", adminHandler.UpdateUpdate)
		admin.DELETE("/updates/:id", adminHandler.DeleteUpdate)
		admin.POST("/updates/:id/rollback", adminHandler.CreateRollback)
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/keys", adminHandler.ListAPIKeys)
		admin.POST("/keys", adminHandler.CreateAPIKey)
		admin.DELETE("/keys/:id", adminHandler.DeleteAPIKey)
	}

	// Print startup banner
	printBanner(cfg.Port)

	// Start self-ping to keep Render free tier awake
	if cfg.Hostname != "" && cfg.Hostname != "http://localhost:8080" {
		go startSelfPing(cfg.Hostname)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getUpdatesDir determines the updates directory location.
func getUpdatesDir() string {
	// Check for explicit UPDATES_DIR env var
	if dir := os.Getenv("UPDATES_DIR"); dir != "" {
		return dir
	}

	// Check relative paths
	candidates := []string{
		"./updates",
		"../updates",
		"../expo-updates-client/dist",
	}

	for _, dir := range candidates {
		if absPath, err := filepath.Abs(dir); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	// Default
	return "./updates"
}

// printBanner prints the server startup banner.
func printBanner(port string) {
	banner := `
  ██████╗ ████████╗ █████╗ ███████╗██╗  ██╗██╗██████╗ 
 ██╔═══██╗╚══██╔══╝██╔══██╗██╔════╝██║  ██║██║██╔══██╗
 ██║   ██║   ██║   ███████║███████╗███████║██║██████╔╝
 ██║   ██║   ██║   ██╔══██║╚════██║██╔══██║██║██╔═══╝ 
 ╚██████╔╝   ██║   ██║  ██║███████║██║  ██║██║██║     
  ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝╚═╝     
                                                                                    
                           OTAShip v%s
                           http://localhost:%s

 Endpoints:
   GET  /api/manifest     - Get update manifest (Expo client)
   GET  /api/assets       - Download assets
   GET  /api/health       - Health check

 Admin API:
   GET  /api/admin/updates        - List updates
   POST /api/admin/updates        - Register update
   PATCH /api/admin/updates/:id   - Modify update
   GET  /api/admin/stats          - Analytics

`
	fmt.Printf(banner, Version, port)
}

// startSelfPing pings the health endpoint every 10 minutes to keep Render free tier awake.
func startSelfPing(hostname string) {
	ticker := time.NewTicker(10 * time.Minute)
	healthURL := hostname + "/api/health"

	log.Printf("Self-ping enabled: %s every 10 minutes", healthURL)

	for range ticker.C {
		resp, err := http.Get(healthURL)
		if err != nil {
			log.Printf("Warning: Self-ping failed: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("Self-ping successful")
	}
}
