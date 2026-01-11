// Package storage provides Cloudinary integration for asset CDN.
package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryService handles asset uploads and URL management.
type CloudinaryService struct {
	cld       *cloudinary.Cloudinary
	cloudName string
	connected bool
	mu        sync.RWMutex
}

// Config holds Cloudinary configuration.
type Config struct {
	CloudName string
	APIKey    string
	APISecret string
}

// Global Cloudinary instance
var Cloudinary *CloudinaryService

// NewCloudinaryService creates a new Cloudinary service.
func NewCloudinaryService(cfg Config) (*CloudinaryService, error) {
	if cfg.CloudName == "" || cfg.APIKey == "" || cfg.APISecret == "" {
		log.Println("Cloudinary not configured, assets will be served locally")
		return nil, nil
	}

	// Create Cloudinary instance
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary client: %w", err)
	}

	service := &CloudinaryService{
		cld:       cld,
		cloudName: cfg.CloudName,
		connected: true,
	}

	Cloudinary = service
	log.Printf("Cloudinary CDN configured (cloud: %s)", cfg.CloudName)

	return service, nil
}

// IsConnected returns true if Cloudinary is configured.
func (s *CloudinaryService) IsConnected() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// UploadResult contains information about an uploaded asset.
type UploadResult struct {
	PublicID    string `json:"publicId"`
	URL         string `json:"url"`
	SecureURL   string `json:"secureUrl"`
	Format      string `json:"format"`
	Size        int    `json:"size"`
	ContentType string `json:"contentType"`
}

// UploadFile uploads a file to Cloudinary.
// folder: the folder path in Cloudinary (e.g., "updates/1/1234567890")
// localPath: path to the local file
// Returns the upload result with Cloudinary URLs.
func (s *CloudinaryService) UploadFile(ctx context.Context, folder, localPath string) (*UploadResult, error) {
	if s == nil || !s.connected {
		return nil, fmt.Errorf("cloudinary not connected")
	}

	// Open the file
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Determine public ID (filename without extension in the folder)
	filename := filepath.Base(localPath)
	ext := filepath.Ext(filename)

	// Determine resource type based on extension
	resourceType := "raw"
	contentType := getContentType(ext)
	if strings.HasPrefix(contentType, "image/") {
		resourceType = "image"
	}

	// Determine public ID
	// For raw files, we must keep the extension so the URL has it.
	// For images, Cloudinary handles extensions, but keeping it is explicit.
	var publicID string
	if resourceType == "raw" {
		publicID = filepath.Join(folder, filename)
	} else {
		publicID = filepath.Join(folder, strings.TrimSuffix(filename, ext))
	}
	publicID = strings.ReplaceAll(publicID, "\\", "/")

	// Upload to Cloudinary
	overwrite := true
	uploadParams := uploader.UploadParams{
		PublicID:     publicID,
		Folder:       "", // Already included in publicID
		ResourceType: resourceType,
		Overwrite:    &overwrite,
	}

	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	return &UploadResult{
		PublicID:    result.PublicID,
		URL:         result.URL,
		SecureURL:   result.SecureURL,
		Format:      result.Format,
		Size:        result.Bytes,
		ContentType: contentType,
	}, nil
}

// UploadDirectory uploads all files in a directory to Cloudinary.
// Returns a map of local paths to Cloudinary URLs.
func (s *CloudinaryService) UploadDirectory(ctx context.Context, folder, localDir string) (map[string]string, error) {
	if s == nil || !s.connected {
		return nil, fmt.Errorf("cloudinary not connected")
	}

	urlMap := make(map[string]string)

	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path for folder structure
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		// Create Cloudinary folder path
		cloudFolder := filepath.Join(folder, filepath.Dir(relPath))
		cloudFolder = strings.ReplaceAll(cloudFolder, "\\", "/")

		// Upload file
		result, err := s.UploadFile(ctx, cloudFolder, path)
		if err != nil {
			log.Printf("Warning: Failed to upload %s: %v", path, err)
			return nil // Continue with other files
		}

		// Store mapping
		urlMap[relPath] = result.SecureURL
		// Store mapping
		urlMap[relPath] = result.SecureURL
		log.Printf("  Uploaded: %s -> %s", relPath, result.SecureURL)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return urlMap, nil
}

// GetAssetURL returns the Cloudinary URL for an asset.
// publicID: the asset's public ID in Cloudinary
func (s *CloudinaryService) GetAssetURL(publicID string) string {
	if s == nil || !s.connected {
		return ""
	}

	// Construct the raw asset URL
	return fmt.Sprintf("https://res.cloudinary.com/%s/raw/upload/%s",
		s.cloudName,
		publicID,
	)
}

// DeleteFolder deletes all assets in a Cloudinary folder.
func (s *CloudinaryService) DeleteFolder(ctx context.Context, folder string) error {
	if s == nil || !s.connected {
		return fmt.Errorf("cloudinary not connected")
	}

	// Delete all assets with the folder prefix
	_, err := s.cld.Admin.DeleteAssetsByPrefix(ctx, admin.DeleteAssetsByPrefixParams{
		Prefix: []string{folder},
	})
	if err != nil {
		return fmt.Errorf("failed to delete assets: %w", err)
	}

	// Delete the empty folder
	_, err = s.cld.Admin.DeleteFolder(ctx, admin.DeleteFolderParams{
		Folder: folder,
	})
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	log.Printf("Deleted Cloudinary folder: %s", folder)
	return nil
}

// getContentType returns the MIME type for a file extension.
func getContentType(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".js", ".hbc":
		return "application/javascript"
	case ".json":
		return "application/json"
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
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}

// ReadFile is a helper to read file content.
func ReadFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
