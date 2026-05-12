package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize OTAship for the current project",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in an Expo project (no app.json found)")
	}

	otashipPath := filepath.Join(projectRoot, "otaship.json")
	if _, err := os.Stat(otashipPath); err == nil {
		return fmt.Errorf("already initialized (otaship.json exists)")
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if cfg.Server == "" {
		ui.Info.Println("No server configured. Running login first...")
		if err := runLogin(cmd, args); err != nil {
			return err
		}

		cfg, _ = config.LoadGlobalConfig()
	}

	fmt.Print("API Key: ")
	var apiKey string
	fmt.Scanln(&apiKey)
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key required")
	}

	c := &client.Client{BaseURL: cfg.Server}
	project, err := c.ValidateAPIKey(apiKey)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}

	ui.Success.Printf("Validated: %s (%s)\n", project.Name, project.Slug)

	cfg.Projects[project.ProjectID] = apiKey
	if err := config.SaveGlobalConfig(cfg); err != nil {
		return err
	}
	ui.Success.Println("Saved API key")

	projectCfg := &config.ProjectConfig{
		ProjectID: project.ProjectID,
		Channel:   "production",
	}
	if err := config.SaveProjectConfig(projectCfg, projectRoot); err != nil {
		return err
	}

	ui.Success.Println("Created otaship.json")
	return nil
}
