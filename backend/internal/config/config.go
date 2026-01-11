// Package config handles application configuration from environment variables.
package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Port     string
	Hostname string

	// MongoDB settings
	MongoDBURI string

	// Cloudinary settings
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string

	// Security settings
	PrivateKeyPath string
	AdminSecret    string
}

// Global application config instance.
var AppConfig *Config

// Load reads configuration from environment variables.
// It first attempts to load from .env file, then falls back to system env vars.
func Load() *Config {
	// Try to load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	AppConfig = &Config{
		Port:                getEnv("PORT", "8080"),
		Hostname:            getEnv("HOSTNAME", "http://localhost:8080"),
		MongoDBURI:          getEnv("MONGODB_URI", ""),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		PrivateKeyPath:      getEnv("PRIVATE_KEY_PATH", "./code-signing-keys/private-key.pem"),
		AdminSecret:         getEnv("ADMIN_SECRET", ""),
	}

	return AppConfig
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

// Validate checks if all required configuration is present.
func (c *Config) Validate() error {
	// Note: MongoDB and Cloudinary are optional during development
	// They become required in production
	if c.MongoDBURI == "" {
		return errors.New("MONGODB_URI is required")
	}
	if c.CloudinaryCloudName == "" {
		return errors.New("CLOUDINARY_CLOUD_NAME is required")
	}
	if c.CloudinaryAPIKey == "" {
		return errors.New("CLOUDINARY_API_KEY is required")
	}
	if c.CloudinaryAPISecret == "" {
		return errors.New("CLOUDINARY_API_SECRET is required")
	}
	return nil
}
