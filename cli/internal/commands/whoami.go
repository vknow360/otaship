package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var WhoAmICmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current project config",
	RunE:  runWhoami,
}

func runWhoami(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("no project config found. Run 'otaship link'")
	}
	gCfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}
	apiKey := gCfg.Projects[cfg.ProjectID]
	if apiKey == "" {
		return fmt.Errorf("no API key found. Run 'otaship link'")
	}
	c := client.Client{BaseURL: gCfg.Server}
	project, err := c.GetProjectByID(apiKey)
	if err != nil {
		return err
	}
	ui.Info.Printf("Project: %s (%s)\n", project.Name, project.Slug)
	return nil
}
