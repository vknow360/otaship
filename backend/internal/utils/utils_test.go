package utils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"
)

func TestProjectIDContext(t *testing.T) {
	ctx := context.Background()
	uuidStr := "123e4567-e89b-12d3-a456-426614174000"
	id, _ := ParseUUID(uuidStr)

	// Test Set and Get
	newCtx := SetProjectId(ctx, id)
	retrievedID := GetProjectId(newCtx)

	if retrievedID.Bytes != id.Bytes {
		t.Errorf("Expected %v, got %v", id.Bytes, retrievedID.Bytes)
	}

	// Test Get with missing value
	missingID := GetProjectId(ctx)
	if missingID.Valid {
		t.Errorf("Expected invalid UUID when not set, got valid")
	}
}

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"Invalid UUID", "invalid-uuid", true},
		{"Empty UUID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseUUID(tt.uuidStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseInt32(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    int32
		wantErr bool
	}{
		{"Valid Positive", "42", 42, false},
		{"Valid Negative", "-42", -42, false},
		{"Invalid String", "abc", 0, true},
		{"Empty String", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt32(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateSHA256(t *testing.T) {
	input := []byte("secret")
	// SHA-256 of "hello world"
	expected := "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b"

	got := CalculateSHA256(input)
	if got != expected {
		t.Errorf("CalculateSHA256() = %v, want %v", got, expected)
	}
}

func TestGetClientIPAndDeviceHash(t *testing.T) {
	req1, _ := http.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.1:1234"

	req2, _ := http.NewRequest("GET", "/", nil)
	req2.Header.Set("x-forwarded-for", "10.0.0.1, 10.0.0.2")

	tests := []struct {
		name     string
		req      *http.Request
		platform string
		wantIP   string
	}{
		{"RemoteAddr Only", req1, "ios", "192.168.1.1"},
		{"X-Forwarded-For", req2, "android", "10.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIP := getClientIP(tt.req)
			if gotIP != tt.wantIP {
				t.Errorf("getClientIP() = %v, want %v", gotIP, tt.wantIP)
			}

			// Test BuildDeviceHash implicitly
			hash := BuildDeviceHash(tt.req, tt.platform)
			if len(hash) != 64 {
				t.Errorf("BuildDeviceHash() returned hash of length %d, want 64", len(hash))
			}
		})
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key1 := GenerateAPIKey()
	key2 := GenerateAPIKey()

	if len(key1) != 64 {
		t.Errorf("GenerateAPIKey() length = %d, want 64 (hex encoding of 32 bytes)", len(key1))
	}

	if key1 == key2 {
		t.Errorf("GenerateAPIKey() returned identical keys: %v", key1)
	}
}

func TestSignManifest(t *testing.T) {
	// Generate a temporary RSA private key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test RSA key: %v", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	manifestData := []byte(`{"id": "123", "version": "1.0.0"}`)

	signature, err := SignManifest(manifestData, string(privateKeyPEM))
	if err != nil {
		t.Fatalf("SignManifest() unexpected error: %v", err)
	}

	if len(signature) == 0 {
		t.Errorf("SignManifest() returned empty signature")
	}

	// Test with invalid key
	_, err = SignManifest(manifestData, "invalid-key-data")
	if err == nil {
		t.Errorf("SignManifest() with invalid key should return error")
	}
}
