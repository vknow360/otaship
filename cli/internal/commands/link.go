package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var LinkCommand = &cobra.Command{
	Use:   "link",
	Short: "Link to existing project (for team members)",
	RunE:  runLink,
}

func runLink(cmd *cobra.Command, args []string) error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load otaship.json: %w", err)
	}
	if projectCfg == nil {
		return fmt.Errorf("no otaship.json found. Run 'otaship init' first")
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

	c := &client.Client{BaseURL: cfg.Server}
	apiKey := cfg.Projects[projectCfg.ProjectID]
	var project *client.ProjectDetails

	if apiKey != "" {
		project, err = c.GetProjectByID(apiKey)
		if err != nil {
			ui.Warning.Printf("Stored API key for project %s is invalid: %v\n", projectCfg.ProjectID, err)
			apiKey = ""
		}
	}

	if apiKey == "" {
		fmt.Print("API Key: ")
		var inputKey string
		fmt.Scanln(&inputKey)
		inputKey = strings.TrimSpace(inputKey)

		if inputKey == "" {
			return fmt.Errorf("API key required")
		}

		validated, err := c.ValidateAPIKey(inputKey)
		if err != nil {
			return fmt.Errorf("invalid API key: %w", err)
		}
		if validated.ProjectID != projectCfg.ProjectID {
			return fmt.Errorf("API key is for different project (Slug: %s, ID: %s)", validated.Slug, validated.ProjectID)
		}

		apiKey = inputKey
		project = &client.ProjectDetails{
			ID:   validated.ProjectID,
			Slug: validated.Slug,
			Name: validated.Name,
		}

		cfg.Projects[project.ID] = apiKey
		if err := config.SaveGlobalConfig(cfg); err != nil {
			return err
		}
		ui.Success.Println("API key saved")
	}

	ui.Success.Printf("Linked to project: %s (%s)\n", project.Name, project.Slug)
	ui.Info.Println("Ready to publish")
	return nil
}
