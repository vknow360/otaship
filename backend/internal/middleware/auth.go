// Package middleware contains HTTP middleware functions.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vknow360/otaship/backend/internal/config"
)

// AdminAuth middleware validates admin API requests.
// Checks for Bearer token in Authorization header.
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.AppConfig

		// Skip auth if no admin secret is configured (development mode)
		if cfg.AdminSecret == "" {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Expected 'Bearer <token>'",
			})
			return
		}

		token := parts[1]

		// Validate token
		if token != cfg.AdminSecret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid admin token",
			})
			return
		}

		c.Next()
	}
}

// CORS middleware adds Cross-Origin Resource Sharing headers.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, expo-platform, expo-runtime-version, expo-channel-name, expo-protocol-version, expo-expect-signature, expo-current-update-id, expo-embedded-update-id")
		c.Header("Access-Control-Expose-Headers", "expo-protocol-version, expo-sfv-version, expo-signature")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestLogger logs incoming requests.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use Gin's built-in logger
		c.Next()
	}
}
