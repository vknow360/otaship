package utils

import (
	"context"
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

	ctx = SetProjectId(ctx, testID)
	got := GetProjectId(ctx)
	if got != testID {
		t.Errorf("GetProjectId() = %v, want %v", got, testID)
	}
}
