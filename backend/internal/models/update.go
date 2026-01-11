// Package models defines the data structures for the application.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Update represents an OTA update bundle stored in the database.
type Update struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectSlug       string             `bson:"projectSlug" json:"projectSlug"`       // Project identifier from app.json expo.slug
	UpdateID          string             `bson:"updateId" json:"updateId"`             // SHA256 UUID format
	RuntimeVersion    string             `bson:"runtimeVersion" json:"runtimeVersion"` // e.g., "1", "2"
	Channel           string             `bson:"channel" json:"channel"`               // production, staging, beta
	Platform          string             `bson:"platform" json:"platform"`             // android, ios, or "all"
	BundlePath        string             `bson:"bundlePath" json:"bundlePath"`         // Local path or Cloudinary folder
	CreatedAt         time.Time          `bson:"createdAt" json:"createdAt"`
	IsActive          bool               `bson:"isActive" json:"isActive"`                   // Can be deactivated
	IsRollback        bool               `bson:"isRollback" json:"isRollback"`               // Rollback directive
	RolloutPercentage int                `bson:"rolloutPercentage" json:"rolloutPercentage"` // 0-100
	Downloads         int64              `bson:"downloads" json:"downloads"`                 // Analytics counter

	// Metadata from the exported bundle
	Metadata *UpdateMetadata `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// UpdateMetadata contains parsed data from metadata.json and expoConfig.json.
type UpdateMetadata struct {
	FileMetadata map[string]PlatformMetadata `bson:"fileMetadata" json:"fileMetadata"`
	ExpoConfig   map[string]interface{}      `bson:"expoConfig,omitempty" json:"expoConfig,omitempty"`
}

// PlatformMetadata contains platform-specific bundle information.
type PlatformMetadata struct {
	Bundle     string  `bson:"bundle" json:"bundle"`
	BundleUrl  string  `bson:"bundleUrl,omitempty" json:"bundleUrl,omitempty"`   // Cloudinary URL
	BundleKey  string  `bson:"bundleKey,omitempty" json:"bundleKey,omitempty"`   // 32-char Hex Hash
	BundleHash string  `bson:"bundleHash,omitempty" json:"bundleHash,omitempty"` // Base64URL Hash (integrity)
	Assets     []Asset `bson:"assets" json:"assets"`
}

// Asset represents a single asset file in the update bundle.
type Asset struct {
	Path string `bson:"path" json:"path"`
	Ext  string `bson:"ext" json:"ext"`
	Url  string `bson:"url,omitempty" json:"url,omitempty"`   // Cloudinary URL
	Key  string `bson:"key,omitempty" json:"key,omitempty"`   // 32-char Hex Hash
	Hash string `bson:"hash,omitempty" json:"hash,omitempty"` // Base64URL Hash (integrity)
}

// CloudinaryAsset maps local assets to Cloudinary URLs.
type CloudinaryAsset struct {
	LocalPath     string `bson:"localPath" json:"localPath"`
	CloudinaryURL string `bson:"cloudinaryUrl" json:"cloudinaryUrl"`
	ContentType   string `bson:"contentType" json:"contentType"`
}

// Constants for channel names.
const (
	ChannelProduction = "production"
	ChannelStaging    = "staging"
	ChannelBeta       = "beta"
)

// Constants for platforms.
const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformAll     = "all"
)
