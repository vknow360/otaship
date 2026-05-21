package handlers

import (
	"reflect"
	"testing"

	"github.com/vknow360/otaship/backend/internal/database"
)

func TestIsLaunchAsset(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{"Valid Launch Asset", "_expo/static/js/ios/AppEntry-1234.js", true},
		{"Bare Launch Asset", "_expo/static/js/App.js", true},
		{"Image Asset", "assets/image.png", false},
		{"Font Asset", "assets/font.ttf", false},
		{"Empty String", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLaunchAsset(tt.fileName)
			if got != tt.expected {
				t.Errorf("isLaunchAsset() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldReceiveUpdate(t *testing.T) {
	tests := []struct {
		name       string
		percentage int
		deviceHash string
		expected   bool
	}{
		{"100 percent always true", 100, "any-hash", true},
		{"0 percent always false", 0, "any-hash", false},
		{"Negative percentage false", -10, "any-hash", false},
		{"Over 100 percentage true", 110, "any-hash", true},
		{
			name:       "Deterministic hash lower than threshold",
			percentage: 50,
			deviceHash: "00000028", // hex for 40 -> 40 < 50 == true
			expected:   true,
		},
		{
			name:       "Deterministic hash higher than threshold",
			percentage: 50,
			deviceHash: "00000050", // hex for 80 -> 80 < 50 == false
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldReceiveUpdate(tt.percentage, tt.deviceHash)
			if got != tt.expected {
				t.Errorf("shouldReceiveUpdate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildAssetsArray(t *testing.T) {
	inputAssets := []database.Asset{
		{
			Hash:     "hash1",
			Key:      "key1",
			MimeType: "image/png",
			Url:      "https://example.com/1.png",
			FileName: "image.png",
		},
		{
			Hash:     "hash2",
			Key:      "key2",
			MimeType: "application/octet-stream",
			Url:      "https://example.com/2",
			FileName: "file_without_extension",
		},
	}

	expectedOutput := []map[string]interface{}{
		{
			"hash":          "hash1",
			"key":           "key1",
			"contentType":   "image/png",
			"url":           "https://example.com/1.png",
			"fileExtension": ".png",
		},
		{
			"hash":        "hash2",
			"key":         "key2",
			"contentType": "application/octet-stream",
			"url":         "https://example.com/2",
			// No fileExtension should be present
		},
	}

	got := buildAssetsArray(inputAssets)
	if !reflect.DeepEqual(got, expectedOutput) {
		t.Errorf("buildAssetsArray() = %v, want %v", got, expectedOutput)
	}
}
