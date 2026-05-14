package commands

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List updates for the current project",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
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

	c := &client.Client{BaseURL: cfg.Server}
	updates, err := c.ListUpdates(apiKey)
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		ui.Info.Println("No updates found")
		return nil
	}

	var tableData [][]string
	tableData = append(tableData, []string{"ID", "PLATFORM", "RUNTIME", "CHANNEL", "ACTIVE", "ROLLOUT", "CREATED"})

	for _, u := range updates {
		active := "✓"
		if !u.IsActive {
			active = "✗"
		}
		if u.IsRollback {
			active = "↩"
		}

		created := time.UnixMilli(u.CreatedAt).Local().Format("2006-01-02 15:04")

		tableData = append(tableData, []string{
			u.ID, u.Platform, u.RuntimeVersion, u.Channel,
			active, fmt.Sprintf("%d", u.RolloutPercentage), created,
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	return nil
}
