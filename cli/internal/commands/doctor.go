package commands

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	_, err := config.LoadProjectConfig()
	if err != nil {
		return err
	}
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	if cfg.Server == "" {
		return fmt.Errorf("no server configured. Run login first")
	}
	ui.Success.Println("Config loaded")
	healthUrl := cfg.Server + "/health"
	resp, err := http.Get(healthUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	ui.Success.Println("Server reachable")
	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in an Expo project (no app.json found)")
	}
	appJson, err := readAppJson(projectRoot)
	if err != nil {
		return err
	}
	if appJson.Expo.RuntimeVersion == "" {
		return fmt.Errorf("no runtimeVersion in app.json")
	}
	ui.Success.Println("Expo project found")
	_, err = exec.LookPath("npx")
	if err != nil {
		return fmt.Errorf("npx not installed")
	}
	ui.Success.Println("npx installed")
	return nil
}
