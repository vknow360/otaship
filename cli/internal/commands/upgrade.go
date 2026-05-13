package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to latest version",
	RunE:  runUpgrade,
}

type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	ui.PrintBanner()
	ui.Info.Printf("Current version: %s\n", Version)

	spinner, _ := ui.StartSpinner("Checking for updates...")
	release, err := fetchLatestRelease()
	if err != nil {
		spinner.Fail("Failed to check for updates")
		return err
	}
	spinner.Success("Update check complete")

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	ui.Info.Printf("Latest version: %s\n", latestVersion)

	if latestVersion == Version {
		ui.Success.Println("Already on latest version!")
		return nil
	}

	ui.Info.Printf("New version available: %s → %s\n", Version, latestVersion)

	if ui.IsInteractive() {
		confirm, err := ui.Confirm("Do you want to upgrade now?")
		if err != nil || !confirm {
			return nil
		}
	}

	asset, err := findAsset(release)
	if err != nil {
		return err
	}

	tmpPath, err := downloadBinary(asset.DownloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	if err := replaceExecutable(tmpPath); err != nil {
		return err
	}

	ui.Success.Printf("Successfully upgraded to v%s!\n", latestVersion)
	ui.Info.Println("Run 'otaship version' to verify.")

	return nil
}

func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	url := "https://api.github.com/repos/vknow360/otaship/releases/latest"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "otaship/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no release found")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func compareVersions(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	return current != latest
}

func findAsset(release *GitHubRelease) (*GitHubAsset, error) {
	assetName := fmt.Sprintf("otaship-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("no binary found for %s", runtime.GOOS)
}

func downloadBinary(downloadURL string) (string, error) {
	spinner, _ := ui.StartSpinner(fmt.Sprintf("Downloading %s...", downloadURL))
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download binary: %d", resp.StatusCode)
	}

	tmpExt := ""
	if runtime.GOOS == "windows" {
		tmpExt = ".exe"
	}

	tmpFile, err := os.CreateTemp("", "otaship-upgrade-*"+tmpExt)
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		spinner.Fail("Download failed")
		return "", err
	}

	spinner.Success("Download complete")
	return tmpPath, nil
}

func replaceExecutable(newPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	execPath, _ = filepath.EvalSymlinks(execPath)

	spinner, _ := ui.StartSpinner("Installing.... ")

	oldPath := execPath + ".old"

	os.Remove(oldPath)

	if err := os.Rename(execPath, oldPath); err != nil {
		return err
	}

	if err := os.Rename(newPath, execPath); err != nil {
		os.Rename(oldPath, execPath)
		spinner.Fail("Install failed")
		return err
	}

	spinner.Success("Install complete")
	return nil
}
