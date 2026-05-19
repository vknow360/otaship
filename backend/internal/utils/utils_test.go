package utils

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCalculateSHA256(t *testing.T) {
	tests := []struct {
		name          string
		inputData     []byte
		expectedValue string
	}{
		{"Test 1", []byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"Test 2", []byte("hello world"), "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValue := CalculateSHA256(tt.inputData)
			if actualValue != tt.expectedValue {
				t.Errorf("CalculateSHA256() = %v, want %v", actualValue, tt.expectedValue)
			}
		})
	}
}

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name          string
		inputData     string
		expectedValue pgtype.UUID
		expectError   bool
	}{
		{
			name:          "Valid UUID",
			inputData:     "123e4567-e89b-12d3-a456-426614174000",
			expectedValue: pgtype.UUID{Bytes: [16]byte{0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00}, Valid: true},
			expectError:   false,
		},
		{
			name:          "Invalid UUID string",
			inputData:     "Hello World",
			expectedValue: pgtype.UUID{},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualValue, err := ParseUUID(tt.inputData)
			if (err != nil) != tt.expectError {
				t.Fatalf("ParseUUID() error = %v, expectError %v", err, tt.expectError)
			}
			if actualValue != tt.expectedValue {
				t.Errorf("ParseUUID() = %v, want %v", actualValue, tt.expectedValue)
			}
		})
	}
}

func TestBuildDeviceHash(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		platform   string
		expected   string
	}{
		{
			name:       "With Port",
			remoteAddr: "192.168.1.5:4000",
			platform:   "android",
			// input should be "192.168.1.5|android"
			expected: CalculateSHA256([]byte("192.168.1.5|android")),
		},
		{
			name:       "Without Port",
			remoteAddr: "10.0.0.1",
			platform:   "ios",
			// input should be "10.0.0.1|ios"
			expected: CalculateSHA256([]byte("10.0.0.1|ios")),
		},
		{
			name:       "Empty IP",
			remoteAddr: "",
			platform:   "android",
			// input should be "unknown|android"
			expected: CalculateSHA256([]byte("unknown|android")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			got := BuildDeviceHash(req, tt.platform)
			if got != tt.expected {
				t.Errorf("BuildDeviceHash() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContextProjectId(t *testing.T) {
	ctx := context.Background()
	testID := pgtype.UUID{Bytes: [16]byte{1, 2, 3}, Valid: true}

	// Test valid
	ctxValid := SetProjectId(ctx, testID)
	got := GetProjectId(ctxValid)
	if got != testID {
		t.Errorf("GetProjectId() valid = %v, want %v", got, testID)
	}

	// Test nil
	gotNil := GetProjectId(ctx)
	if gotNil.Valid {
		t.Errorf("GetProjectId() nil should return invalid pgtype.UUID")
	}

	// Test invalid type
	ctxInvalid := context.WithValue(ctx, projectIDKey, "not-a-uuid")
	gotInvalid := GetProjectId(ctxInvalid)
	if gotInvalid.Valid {
		t.Errorf("GetProjectId() invalid type should return invalid pgtype.UUID")
	}
}

func TestParseInt32(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int32
		expectError bool
	}{
		{"Valid positive", "123", 123, false},
		{"Valid negative", "-456", -456, false},
		{"Invalid string", "abc", 0, true},
		{"Empty string", "", 0, true},
		{"Overflow", "2147483648", 0, true}, // MaxInt32 + 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt32(tt.input)
			if (err != nil) != tt.expectError {
				t.Fatalf("ParseInt32() error = %v, expectError %v", err, tt.expectError)
			}
			if got != tt.expected {
				t.Errorf("ParseInt32() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"x-forwarded-for": "203.0.113.195"},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"x-forwarded-for": "203.0.113.195, 150.172.238.178"},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "RemoteAddr with port",
			headers:    map[string]string{},
			remoteAddr: "198.51.100.1:443",
			expected:   "198.51.100.1",
		},
		{
			name:       "RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "198.51.100.1",
			expected:   "198.51.100.1",
		},
		{
			name:       "Empty values",
			headers:    map[string]string{},
			remoteAddr: "",
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remoteAddr

			got := getClientIP(req)
			if got != tt.expected {
				t.Errorf("getClientIP() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key1 := GenerateAPIKey()
	key2 := GenerateAPIKey()

	if key1 == key2 {
		t.Error("GenerateAPIKey() generated duplicate keys")
	}

	if len(key1) != 64 {
		t.Errorf("GenerateAPIKey() length = %d, want 64", len(key1))
	}
}

func TestSignManifest(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test RSA key: %v", err)
	}

	manifest := []byte(`{"version":"1.0"}`)

	// PKCS8 format (standard test)
	privBytesPKCS8, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	pemPKCS8 := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytesPKCS8}))

	signature, err := SignManifest(manifest, pemPKCS8)
	if err != nil {
		t.Fatalf("SignManifest() error = %v", err)
	}
	if signature == "" {
		t.Error("SignManifest() returned empty signature")
	}

	// PKCS1 format
	privBytesPKCS1 := x509.MarshalPKCS1PrivateKey(privateKey)
	pemPKCS1 := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytesPKCS1}))
	_, err = SignManifest(manifest, pemPKCS1)
	if err != nil {
		t.Fatalf("SignManifest() PKCS1 error = %v", err)
	}

	// Invalid key format (fails both PKCS8 and PKCS1 parsing)
	invalidBytes := []byte("invalid-key-bytes")
	pemInvalid := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: invalidBytes}))
	_, err = SignManifest(manifest, pemInvalid)
	if err == nil {
		t.Error("SignManifest() expected error for unparseable key")
	}

	// Not an RSA key (e.g. ECDSA)
	ecdsaKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecdsaBytes, _ := x509.MarshalPKCS8PrivateKey(ecdsaKey)
	pemECDSA := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecdsaBytes}))
	_, err = SignManifest(manifest, pemECDSA)
	if err == nil || err.Error() != "parsed key is not an RSA private key" {
		t.Errorf("SignManifest() expected 'not an RSA private key' error, got: %v", err)
	}

	// Empty string
	_, err = SignManifest(manifest, "invalid key")
	if err == nil {
		t.Error("SignManifest() expected error for invalid PEM string")
	}
}
