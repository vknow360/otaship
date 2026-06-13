package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadProjectConfig(t *testing.T) {
	// Create a temp directory to act as our project root
	tempDir, err := os.MkdirTemp("", "otaship-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// We can't easily test FindProjectRoot without manipulating the filesystem heavily,
	// but we can test SaveProjectConfig and the internal loadJSON logic directly.

	cfg := &ProjectConfig{
		ProjectID: "test-project-123",
		Channel:   "staging",
	}

	err = SaveProjectConfig(cfg, tempDir)
	if err != nil {
		t.Fatalf("SaveProjectConfig failed: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, "otaship.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected otaship.json to be created, but it was not")
	}

	// Load the config back using loadJSON directly since LoadProjectConfig relies on FindProjectRoot
	var loadedCfg ProjectConfig
	err = loadJSON(configPath, &loadedCfg)
	if err != nil {
		t.Fatalf("loadJSON failed: %v", err)
	}

	if loadedCfg.ProjectID != cfg.ProjectID {
		t.Errorf("Expected ProjectID %s, got %s", cfg.ProjectID, loadedCfg.ProjectID)
	}
	if loadedCfg.Channel != cfg.Channel {
		t.Errorf("Expected Channel %s, got %s", cfg.Channel, loadedCfg.Channel)
	}
}

func TestLoadProjectConfig_MissingProjectID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "otaship-test-missing-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &ProjectConfig{
		Channel: "staging", // Missing ProjectID
	}

	configPath := filepath.Join(tempDir, "otaship.json")
	err = saveJSONAtomic(configPath, cfg)
	if err != nil {
		t.Fatalf("saveJSONAtomic failed: %v", err)
	}

	// Read it back
	var loadedCfg ProjectConfig
	err = loadJSON(configPath, &loadedCfg)
	if err != nil {
		t.Fatalf("loadJSON failed: %v", err)
	}

	if loadedCfg.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", loadedCfg.ProjectID)
	}
}
