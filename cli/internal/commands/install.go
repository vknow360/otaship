package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install OTAShip CLI",
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	ui.PrintBanner()

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve path: %w", err)
	}

	ui.Info.Printf("Current location: %s\n", execPath)

	installPath, err := getInstallPath()
	if err != nil {
		return fmt.Errorf("could not determine install path: %w", err)
	}

	if execPath == installPath {
		ui.Info.Println("Already installed")
		return nil
	}

	ui.Info.Printf("Will install to: %s\n", installPath)

	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("could not create install directory: %w", err)
	}

	spinner, _ := ui.StartSpinner("Installing binary...")
	if err := copyFile(execPath, installPath); err != nil {
		spinner.Fail("Failed to copy binary")
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	spinner.Success("Installation complete!")

	installDir = filepath.Dir(installPath)

	if runtime.GOOS == "windows" {
		ui.Warning.Println("To use from anywhere, add to PATH:")
		ui.Info.Printf("  setx PATH \"%%PATH%%;%s\"\n", installDir)
		ui.Info.Println("  (Restart terminal after)")
	} else {
		shell := os.Getenv("SHELL")
		rcFile := "~/.bashrc"
		if strings.Contains(shell, "zsh") {
			rcFile = "~/.zshrc"
		}

		ui.Warning.Println("To use from anywhere, add to PATH:")
		ui.Info.Printf("  echo 'export PATH=\"%s:$PATH\"' >> %s\n", installDir, rcFile)
		ui.Info.Printf("  source %s\n", rcFile)
	}

	return nil
}

func getInstallPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var installDir string
	var binaryName string

	switch runtime.GOOS {
	case "windows":
		installDir = filepath.Join(homeDir, "AppData", "Roaming", "otaship")
		binaryName = "otaship.exe"
	case "darwin", "linux":
		installDir = filepath.Join(homeDir, ".local", "bin")
		binaryName = "otaship"
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return filepath.Join(installDir, binaryName), nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
