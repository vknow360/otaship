package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [update-id]",
	Short: "Delete an update",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	updateID := args[0]

	projectCfg, err := config.LoadProjectConfig()
	if err != nil || projectCfg == nil {
		return fmt.Errorf("not in an OTAship project. Run 'otaship init'")
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	apiKey := cfg.Projects[projectCfg.ProjectID]
	if apiKey == "" {
		return fmt.Errorf("no API key found. Run 'otaship link'")
	}

	spinner, _ := ui.StartSpinner(fmt.Sprintf("Deleting update %s...", updateID))

	c := &client.Client{BaseURL: cfg.Server}
	err = c.DeleteUpdate(apiKey, updateID)
	if err != nil {
		spinner.Fail("FAILED")
		return err
	}

	spinner.Success(fmt.Sprintf("Update %s deleted", updateID))
	return nil
}
