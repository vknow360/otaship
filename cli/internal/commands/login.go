package commands

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set OTAShip server URL",
	RunE:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Server != "" {
		ui.Info.Printf("Current server: %s\n", cfg.Server)
		change, err := ui.Confirm("Change server?")
		if err != nil || !change {
			return nil
		}
	}

	server, err := ui.Ask("Enter server URL")
	if err != nil {
		return err
	}
	return setServer(cfg, server)
}

func setServer(cfg *config.GlobalConfig, serverUrl string) error {
	serverUrl = strings.TrimSpace(serverUrl)

	parsed, err := url.Parse(serverUrl)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid URL format: ensure it includes http:// or https://")
	}

	if err := pingServer(serverUrl); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	cfg.Server = serverUrl
	if err := config.SaveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	ui.Success.Printf("Server set to %s\n", serverUrl)
	return nil
}

func pingServer(serverURL string) error {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}
