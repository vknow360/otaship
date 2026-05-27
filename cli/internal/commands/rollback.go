package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var RollbackCmd = &cobra.Command{
	Use:   "rollback [update-id]",
	Short: "Republish a specific older update",
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

func runRollback(cmd *cobra.Command, args []string) error {
	updateID := args[0]

	projectCfg, err := config.LoadProjectConfig()
	if err != nil || projectCfg == nil {
		return fmt.Errorf("not in an OTAShip project. Run 'otaship init'")
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	apiKey := cfg.Projects[projectCfg.ProjectID]
	if apiKey == "" {
		return fmt.Errorf("no API key found. Run 'otaship link'")
	}

	spinner, _ := ui.StartSpinner(fmt.Sprintf("Rolling back update %s...", updateID))

	c := &client.Client{BaseURL: cfg.Server}
	rollback, err := c.RollbackUpdate(apiKey, updateID)
	if err != nil {
		spinner.Fail("FAILED")
		return err
	}

	spinner.Success(fmt.Sprintf("Rollback created: %s", rollback.ID))
	ui.Info.Println("This older update has been republished and is now active")
	return nil
}
