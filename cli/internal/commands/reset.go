package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Rollback devices to the embedded update (factory reset)",
	Args:  cobra.NoArgs,
	RunE:  runReset,
}

func init() {
	ResetCmd.Flags().StringVar(&platformFlag, "platform", "all", "Platform: android, ios, or all")
	ResetCmd.Flags().StringVar(&channelFlag, "channel", "", "Override channel (default from config)")
}

func runReset(cmd *cobra.Command, args []string) error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil || projectCfg == nil {
		return fmt.Errorf("not in an OTAShip project. Run 'otaship init'")
	}

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in an Expo project (no app.json found)")
	}

	appJson, err := readAppJson(projectRoot)
	if err != nil {
		return err
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	apiKey := cfg.Projects[projectCfg.ProjectID]
	if apiKey == "" {
		return fmt.Errorf("no API key found. Run 'otaship link'")
	}

	platform, _ := resolvePlatform(cmd)
	channel, _ := resolveChannel(cmd, projectCfg.Channel)

	if ui.IsInteractive() && !yesFlag {
		ui.Info.Println(fmt.Sprintf("This will reset %s devices on channel '%s' (Runtime: %s) to their factory built-in updates.", platform, channel, appJson.Expo.RuntimeVersion))
		confirmed, err := ui.Confirm("Proceed?")
		if err != nil || !confirmed {
			return fmt.Errorf("reset cancelled")
		}
	}

	c := &client.Client{BaseURL: cfg.Server}

	resetPlatform := func(p string) error {
		spinner, _ := ui.StartSpinner(fmt.Sprintf("Triggering reset for %s...", p))

		req := &client.RollbackToEmbeddedRequest{
			Platform:       p,
			RuntimeVersion: appJson.Expo.RuntimeVersion,
			Channel:        channel,
		}

		_, err := c.RollbackToEmbedded(apiKey, projectCfg.ProjectID, req)
		if err != nil {
			spinner.Fail(fmt.Sprintf("Failed to reset %s", p))
			return err
		}

		spinner.Success(fmt.Sprintf("Reset created for %s", p))
		return nil
	}

	if platform == "all" || platform == "android" {
		if err := resetPlatform("android"); err != nil {
			return err
		}
	}

	if platform == "all" || platform == "ios" {
		if err := resetPlatform("ios"); err != nil {
			return err
		}
	}

	ui.Success.Println("Devices will revert to the built-in update on their next check")
	return nil
}
