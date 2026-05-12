package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type GlobalConfig struct {
	Version  string            `json:"version"`
	Server   string            `json:"server"`
	Projects map[string]string `json:"projects"`
}

type ProjectConfig struct {
	ProjectID string `json:"project_id"`
	Channel   string `json:"channel"`
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}
	var cfg GlobalConfig
	err = loadJSON(path, &cfg)
	if os.IsNotExist(err) {
		return &GlobalConfig{Version: "1", Projects: make(map[string]string)}, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveGlobalConfig(cfg *GlobalConfig) error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return saveJSONAtomic(path, cfg)
}

func LoadProjectConfig() (*ProjectConfig, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	configFilePath := filepath.Join(projectRoot, "otaship.json")
	err = loadJSON(configFilePath, &cfg)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("invalid otaship.json: missing project_id")
	}

	if cfg.Channel == "" {
		cfg.Channel = "production" // Default
	}

	return &cfg, nil
}

func SaveProjectConfig(cfg *ProjectConfig, projectRoot string) error {
	return saveJSONAtomic(filepath.Join(projectRoot, "otaship.json"), cfg)
}

func loadJSON(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}
func saveJSONAtomic(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	temp, err := os.CreateTemp(dir, "config-*.json")
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(temp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(temp.Name(), 0600); err != nil {
		return err
	}

	if err := os.Rename(temp.Name(), path); err != nil {
		return err
	}
	return nil
}
