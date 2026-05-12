package commands

import (
	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var StatusCommand = &cobra.Command{
	Use:   "status",
	Short: "Show current project status",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil || projectCfg == nil {
		return err
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}
	c := &client.Client{
		BaseURL: cfg.Server,
	}
	project, err := c.GetProjectByID(cfg.Projects[projectCfg.ProjectID])
	if err != nil {
		return err
	}

	hasKey := cfg.Projects[projectCfg.ProjectID] != ""
	ui.Info.Printf("Project: %s (%s)\n", project.Name, project.Slug)
	ui.Info.Printf("Project ID: %s\n", projectCfg.ProjectID)
	ui.Info.Printf("Server: %s\n", cfg.Server)
	ui.Info.Printf("Channel: %s\n", projectCfg.Channel)

	if hasKey {
		ui.Success.Println("API Key: Configured")
	} else {
		ui.Warning.Println("API Key: Missing (run 'otaship link')")
	}
	return nil
}
