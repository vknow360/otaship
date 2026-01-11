// Package handlers contains HTTP request handlers.
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vknow360/otaship/backend/internal/config"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/services"
)

// ManifestHandler handles the /api/manifest endpoint.
// This is the primary endpoint that Expo clients call to check for updates.
type ManifestHandler struct {
	signingService *services.SigningService
	rolloutService *services.RolloutService
	updateRepo     *database.UpdateRepository
	updatesDir     string
}

// NewManifestHandler creates a new manifest handler.
func NewManifestHandler(
	signingService *services.SigningService,
	rolloutService *services.RolloutService,
	updateRepo *database.UpdateRepository,
	updatesDir string,
) *ManifestHandler {
	return &ManifestHandler{
		signingService: signingService,
		rolloutService: rolloutService,
		updateRepo:     updateRepo,
		updatesDir:     updatesDir,
	}
}

// Handle processes manifest requests from Expo clients.
func (h *ManifestHandler) Handle(c *gin.Context) {
	// Only GET requests allowed
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Expected GET."})
		return
	}

	// Parse protocol version
	protocolVersionStr := c.GetHeader("expo-protocol-version")
	if protocolVersionStr == "" {
		protocolVersionStr = "0"
	}
	protocolVersion, _ := strconv.Atoi(protocolVersionStr)

	// Parse platform
	platform := c.GetHeader("expo-platform")
	if platform == "" {
		platform = c.Query("platform")
	}
	if platform != "ios" && platform != "android" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported platform. Expected either ios or android.",
		})
		return
	}

	// Parse runtime version
	runtimeVersion := c.GetHeader("expo-runtime-version")
	if runtimeVersion == "" {
		runtimeVersion = c.Query("runtime-version")
	}
	if runtimeVersion == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No runtimeVersion provided."})
		return
	}

	// Parse channel (defaults to production)
	channel := c.GetHeader("expo-channel-name")
	if channel == "" {
		channel = "production"
	}

	// Get current update ID
	currentUpdateID := c.GetHeader("expo-current-update-id")

	// Get project slug from URL
	projectSlug := c.Param("projectSlug")
	if projectSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	log.Printf("Manifest request: project=%s runtime=%s platform=%s channel=%s", projectSlug, runtimeVersion, platform, channel)

	// 1. Find latest update in DB
	if h.updateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not connected"})
		return
	}

	update, err := h.updateRepo.FindLatest(c.Request.Context(), projectSlug, runtimeVersion, channel, platform)
	if err != nil || update == nil {
		h.handleNoUpdateAvailable(c, protocolVersion)
		return
	}

	// 2. Check if it's a rollback
	if update.IsRollback {
		embeddedUpdateID := c.GetHeader("expo-embedded-update-id")
		if currentUpdateID == embeddedUpdateID {
			h.handleNoUpdateAvailable(c, protocolVersion)
			return
		}

		directive := map[string]interface{}{
			"type": "rollBackToEmbedded",
			"parameters": map[string]interface{}{
				"commitTime": update.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
			},
		}
		directiveJSON, _ := json.Marshal(directive)
		var signature string
		if c.GetHeader("expo-expect-signature") != "" && h.signingService.IsLoaded() {
			signature, _ = h.signingService.CreateSignatureHeader(directiveJSON)
		}
		h.sendMultipartResponse(c, "directive", directiveJSON, signature, protocolVersion, config.AppConfig)
		return
	}

	// 3. Check if client already has this update
	if currentUpdateID == update.UpdateID && protocolVersion == 1 {
		h.handleNoUpdateAvailable(c, protocolVersion)
		return
	}

	// 4. Prepare metadata
	if update.Metadata == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update metadata missing"})
		return
	}

	metadataBytes, _ := json.Marshal(update.Metadata)
	var metadata map[string]interface{}
	json.Unmarshal(metadataBytes, &metadata)

	createdAt := update.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z")

	// Increment downloads - MOVED TO ASSETS HANDLER
	// _ = h.updateRepo.IncrementDownloads(c.Request.Context(), update.ID.Hex())

	// 5. Send manifest
	h.sendManifest(c, update.BundlePath, metadata, update.UpdateID, update.ID.Hex(), createdAt, runtimeVersion, platform, protocolVersion)
}

// sendManifest builds and sends the manifest response.
func (h *ManifestHandler) sendManifest(
	c *gin.Context,
	updateBundlePath string,
	metadata map[string]interface{},
	updateID, updateObjectID, createdAt, runtimeVersion, platform string,
	protocolVersion int,
) {
	cfg := config.AppConfig

	// Get platform-specific metadata
	fileMetadata, ok := metadata["fileMetadata"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid metadata format"})
		return
	}

	platformMetadata, ok := fileMetadata[platform].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No metadata for platform"})
		return
	}

	// Build assets list
	assets := []map[string]interface{}{}
	if assetsList, ok := platformMetadata["assets"].([]interface{}); ok {
		for _, a := range assetsList {
			asset := a.(map[string]interface{})
			assetPath := asset["path"].(string)
			ext := asset["ext"].(string)
			url, _ := asset["url"].(string)
			key, _ := asset["key"].(string)
			hash, _ := asset["hash"].(string)

			// Helper to get asset info (DB prioritized)
			assetInfo, err := h.getAssetInfoFromMetadataOrFile(updateBundlePath, assetPath, ext, key, hash, url, "", "", runtimeVersion, platform, false)
			if err == nil {
				assets = append(assets, assetInfo)
			}
		}
	}

	// Build launch asset
	bundlePath := platformMetadata["bundle"].(string)
	bundleUrl, _ := platformMetadata["bundleUrl"].(string)
	bundleKey, _ := platformMetadata["bundleKey"].(string)
	bundleHash, _ := platformMetadata["bundleHash"].(string)

	launchAsset, err := h.getAssetInfoFromMetadataOrFile(updateBundlePath, bundlePath, "bundle", bundleKey, bundleHash, bundleUrl, updateObjectID, "true", runtimeVersion, platform, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get launch asset"})
		return
	}

	// Get expo config (from metadata if available, else file)
	var expoConfig map[string]interface{}
	if val, ok := metadata["expoConfig"]; ok && val != nil {
		expoConfig = val.(map[string]interface{})
	} else {
		// Fallback to legacy file read (might fail if deleted)
		expoConfig, _ = h.getExpoConfig(updateBundlePath)
	}

	// Build manifest
	manifest := map[string]interface{}{
		"id":             updateID,
		"createdAt":      createdAt,
		"runtimeVersion": runtimeVersion,
		"assets":         assets,
		"launchAsset":    launchAsset,
		"metadata":       map[string]interface{}{},
		"extra": map[string]interface{}{
			"expoClient": expoConfig,
		},
	}

	manifestJSON, _ := json.Marshal(manifest)

	// Sign manifest if requested
	var signature string
	if c.GetHeader("expo-expect-signature") != "" && h.signingService.IsLoaded() {
		signature, _ = h.signingService.CreateSignatureHeader(manifestJSON)
	}

	// Build multipart response
	h.sendMultipartResponse(c, "manifest", manifestJSON, signature, protocolVersion, cfg)
}

// getAssetInfoFromMetadataOrFile builds asset metadata, preferring DB metadata.
func (h *ManifestHandler) getAssetInfoFromMetadataOrFile(
	updateBundlePath, assetPath, ext, key, hash, url, updateObjectID, isLaunchAssetParam, runtimeVersion, platform string,
	isLaunchAsset bool,
) (map[string]interface{}, error) {
	cfg := config.AppConfig

	contentType := "application/octet-stream"
	if isLaunchAsset {
		contentType = "application/javascript"
	} else {
		switch ext {
		case "png":
			contentType = "image/png"
		case "jpg", "jpeg":
			contentType = "image/jpeg"
		case "gif":
			contentType = "image/gif"
		case "svg":
			contentType = "image/svg+xml"
		case "ttf":
			contentType = "font/ttf"
		case "otf":
			contentType = "font/otf"
		case "woff":
			contentType = "font/woff"
		case "woff2":
			contentType = "font/woff2"
		}
	}

	fileExt := "." + ext
	if isLaunchAsset {
		fileExt = ".bundle"
	}

	// Logic: We ALWAYS want to route through our tracking URL if we want to count.
	// But counting is mainly important for BUNDLE (launch asset).
	// If it's a regular asset, we can direct link to Cloudinary (optional) or proxy.
	// Let's proxy everything for consistency or just proxy bundle?
	// User said "only increment when download api is called".

	// Construct the Local/Proxy URL
	// Note: If Url (Cloudinary) exists, we embed it as redirect param.

	relPath := assetPath // Default
	if filepath.IsAbs(assetPath) {
		p, err := filepath.Rel(h.updatesDir, assetPath)
		if err == nil {
			relPath = p
		}
	}

	baseAssetURL := fmt.Sprintf("%s/api/assets?asset=%s&runtimeVersion=%s&platform=%s",
		cfg.Hostname,
		filepath.ToSlash(relPath),
		runtimeVersion,
		platform,
	)

	if url != "" {
		baseAssetURL += fmt.Sprintf("&redirect=%s", url)
	}
	if isLaunchAssetParam == "true" {
		baseAssetURL += "&isLaunchAsset=true"
	}
	if updateObjectID != "" {
		baseAssetURL += fmt.Sprintf("&updateId=%s", updateObjectID)
	}

	// If DB metadata present, use tracking URL for download counting
	if key != "" && hash != "" {
		return map[string]interface{}{
			"hash":          hash,
			"key":           key,
			"fileExtension": fileExt,
			"contentType":   contentType,
			"url":           baseAssetURL,
		}, nil
	}

	// Fallback to file reading
	fullPath := filepath.Join(updateBundlePath, assetPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	computedHash := services.Base64URLEncode(services.ComputeSHA256HashBytes(data))
	computedKey := services.ComputeSHA256Hash(data)[:32]

	return map[string]interface{}{
		"hash":          computedHash,
		"key":           computedKey,
		"fileExtension": fileExt,
		"contentType":   contentType,
		"url":           baseAssetURL,
	}, nil
}

// getExpoConfig reads the expoConfig.json file.
func (h *ManifestHandler) getExpoConfig(updateBundlePath string) (map[string]interface{}, error) {
	configPath := filepath.Join(updateBundlePath, "expoConfig.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// sendMultipartResponse sends the multipart/mixed response.
func (h *ManifestHandler) sendMultipartResponse(
	c *gin.Context,
	partName string,
	data []byte,
	signature string,
	protocolVersion int,
	cfg *config.Config,
) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", "application/json")
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, partName))
	if signature != "" {
		partHeader.Set("expo-signature", signature)
	}

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create response"})
		return
	}

	part.Write(data)
	writer.Close()

	c.Header("expo-protocol-version", strconv.Itoa(protocolVersion))
	c.Header("expo-sfv-version", "0")
	c.Header("cache-control", "private, max-age=0")
	c.Data(http.StatusOK, "multipart/mixed; boundary="+writer.Boundary(), buf.Bytes())
}

// handleNoUpdateAvailable sends the "no update available" directive.
func (h *ManifestHandler) handleNoUpdateAvailable(c *gin.Context, protocolVersion int) {
	if protocolVersion == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No update available"})
		return
	}

	directive := map[string]interface{}{
		"type": "noUpdateAvailable",
	}

	directiveJSON, _ := json.Marshal(directive)

	var signature string
	if c.GetHeader("expo-expect-signature") != "" && h.signingService.IsLoaded() {
		signature, _ = h.signingService.CreateSignatureHeader(directiveJSON)
	}

	h.sendMultipartResponse(c, "directive", directiveJSON, signature, protocolVersion, config.AppConfig)
}
