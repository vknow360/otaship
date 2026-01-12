// Package services contains business logic for the application.
package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
)

// SigningService handles RSA-SHA256 code signing for manifests.
type SigningService struct {
	privateKey *rsa.PrivateKey
	mu         sync.RWMutex
}

// Global signing service instance.
var Signer *SigningService

// NewSigningService creates a new signing service.
func NewSigningService() *SigningService {
	return &SigningService{}
}

// LoadPrivateKey loads the RSA private key from the specified path.
func (s *SigningService) LoadPrivateKey(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read private key file
	pemData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format as fallback
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not an RSA key")
	}

	s.privateKey = rsaKey
	return nil
}

// IsLoaded returns true if a private key has been loaded.
func (s *SigningService) IsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.privateKey != nil
}

// SignManifest signs the manifest data using RSA-SHA256.
// Returns base64-encoded signature.
func (s *SigningService) SignManifest(data []byte) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.privateKey == nil {
		return "", fmt.Errorf("private key not loaded")
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(data)

	// Sign with RSA PKCS1v15
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// CreateSignatureHeader creates the signature header value for expo-signature.
// Format: sig=<base64-signature>, keyid="main"
func (s *SigningService) CreateSignatureHeader(data []byte) (string, error) {
	sig, err := s.SignManifest(data)
	if err != nil {
		return "", err
	}

	// Format as structured field dictionary
	return fmt.Sprintf(`sig="%s", keyid="main"`, sig), nil
}

// ComputeSHA256Hash computes and returns the SHA256 hash of data.
func ComputeSHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// ComputeSHA256HashBytes computes and returns the raw SHA256 hash specific bytes.
func ComputeSHA256HashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ConvertSHA256ToUUID converts a SHA256 hash to UUID format.
// Used for update IDs.
func ConvertSHA256ToUUID(hash string) string {
	if len(hash) < 32 {
		return hash
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hash[0:8],
		hash[8:12],
		hash[12:16],
		hash[16:20],
		hash[20:32],
	)
}

// Base64URLEncode encodes data using URL-safe base64 without padding.
// Base64URLEncode encodes data using URL-safe base64 without padding.
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
