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
	Short: "Set OTAship server URL",
	RunE:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Server != "" {
		ui.Info.Printf("Current server: %s\n", cfg.Server)
		fmt.Print("Change server? (y/n/URL): ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		if response == "n" {
			return nil
		}
		if response != "y" && (strings.HasPrefix(response, "http://") || strings.HasPrefix(response, "https://")) {
			// User entered the URL directly
			return setServer(cfg, response)
		}
		if response != "y" {
			return nil
		}
	}

	fmt.Print("Enter server URL: ")
	var server string
	_, err = fmt.Scanln(&server)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
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
