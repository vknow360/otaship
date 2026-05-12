package config

import (
	"os"
	"path/filepath"
)

// GetGlobalConfigPath returns the full path to ~/.otaship/config.json
func GetGlobalConfigPath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHome, ".otaship", "config.json"), nil
}

func GetGlobalConfigDir() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHome, ".otaship"), nil
}

// CreateConfigDir creates ~/.otaship/ directory if it doesn't exist
func CreateConfigDir() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDirPath := filepath.Join(userHome, ".otaship")
	return os.MkdirAll(configDirPath, 0755)
}

// FindProjectRoot searches upward from current dir for otaship.json
// Returns the directory containing otaship.json, not the file path
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	home, _ := os.UserHomeDir()

	for {
		configFilePath := filepath.Join(dir, "app.json")

		if _, err := os.Stat(configFilePath); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if dir == home || parentDir == dir {
			break
		}
		dir = parentDir
	}
	return "", os.ErrNotExist
}
