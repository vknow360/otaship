package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vknow360/otaship/cli/internal/client"
	"github.com/vknow360/otaship/cli/internal/config"
	"github.com/vknow360/otaship/cli/internal/ui"
)

var (
	channelFlag  string
	rolloutFlag  int
	skipExport   bool
	platformFlag string
	dryRunFlag   bool
)

var PublishCommand = &cobra.Command{
	Use:   "publish",
	Short: "Publish OTA update",
	RunE:  runPublish,
}

type AppJson struct {
	Expo struct {
		Slug           string `json:"slug"`
		RuntimeVersion string `json:"runtimeVersion"`
	} `json:"expo"`
}

func init() {
	PublishCommand.Flags().StringVar(&channelFlag, "channel", "", "Override channel (default from config)")
	PublishCommand.Flags().IntVar(&rolloutFlag, "rollout", 100, "Rollout percentage (0-100)")
	PublishCommand.Flags().BoolVar(&skipExport, "skip-export", false, "Skip expo export step")
	PublishCommand.Flags().StringVar(&platformFlag, "platform", "all", "Platform: android, ios, or all")
	PublishCommand.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry run (no actual update)")
}

func runPublish(cmd *cobra.Command, args []string) error {
	projectCfg, err := config.LoadProjectConfig()
	if err != nil || projectCfg == nil {
		return fmt.Errorf("not in an OTAship project. Run 'otaship init'")
	}

	if channelFlag != "" {
		projectCfg.Channel = channelFlag
	}

	if rolloutFlag < 0 || rolloutFlag > 100 {
		return fmt.Errorf("rollout percentage must be between 0 and 100")
	}

	if platformFlag != "android" && platformFlag != "ios" && platformFlag != "all" {
		return fmt.Errorf("invalid platform. Expected android, ios, or all")
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	c := &client.Client{BaseURL: cfg.Server}
	project, err := c.GetProjectByID(cfg.Projects[projectCfg.ProjectID])
	if err != nil {
		return err
	}

	apiKey, ok := cfg.Projects[project.ID]
	if !ok || apiKey == "" {
		return fmt.Errorf("no API key found. Run 'otaship link'")
	}

	ui.PrintBanner()
	ui.Info.Printf("Publishing to: %s\n", project.Name)
	ui.Info.Printf("Channel: %s\n", projectCfg.Channel)

	projectRoot, err := config.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in an Expo project (no app.json found)")
	}

	if !skipExport {
		defer os.RemoveAll(filepath.Join(projectRoot, "dist"))
	}

	appJson, err := readAppJson(projectRoot)
	if err != nil {
		return err
	}
	ui.Success.Printf("Project: %s\n", project.Name)
	ui.Success.Printf("Runtime: %s\n", appJson.Expo.RuntimeVersion)

	publish := func(platform string, updateReq *client.CreateUpdateRequest) error {
		if dryRunFlag {
			ui.Info.Printf("[DRY RUN] Would create update for %s (Runtime: %s, Channel: %s)\n",
				platform, updateReq.RuntimeVersion, updateReq.Channel)
		} else {
			update, err := c.CreateUpdate(apiKey, updateReq)
			if err != nil {
				return fmt.Errorf("failed to create update: %w", err)
			}
			ui.Success.Printf("Created update: %s\n", update.ID)
			updateReq.ID = update.ID
		}

		spinner, _ := ui.StartSpinner(fmt.Sprintf("Packaging %s bundle...", platform))
		bundleZip, err := zipDistFolder(projectRoot, platform)
		if err != nil {
			spinner.Fail(fmt.Sprintf("Failed to package %s bundle", platform))
			return err
		}
		spinner.Success(fmt.Sprintf("Packaged %s bundle", platform))
		defer os.Remove(bundleZip)

		if dryRunFlag {
			fi, _ := os.Stat(bundleZip)
			ui.Info.Printf("[DRY RUN] Would upload %s bundle (Size: %s)\n", platform, formatSize(fi.Size()))
			ui.Success.Printf("[DRY RUN] %s publish simulated successfully\n", platform)
			ui.Info.Printf("Update ID: [DRY-RUN]\n")
		} else {
			spinner, _ = ui.StartSpinner(fmt.Sprintf("Uploading %s bundle...", platform))
			if err := c.UploadBundle(projectCfg.ProjectID, updateReq.ID, platform, apiKey, bundleZip); err != nil {
				spinner.Fail(fmt.Sprintf("%s upload failed", platform))
				return fmt.Errorf("%s upload failed: %w", platform, err)
			}
			spinner.Success(fmt.Sprintf("Uploaded %s bundle", platform))
			ui.Success.Println("Published successfully!")
			ui.Info.Printf("Update ID: %s\n", updateReq.ID)
		}
		ui.Info.Printf("Channel: %s\n", projectCfg.Channel)
		return nil
	}

	if platformFlag == "all" || platformFlag == "android" {
		if !skipExport {
			spinner, _ := ui.StartSpinner("Running expo export for android...")
			err = runExpoExport(projectRoot, "android")
			if err != nil {
				spinner.Fail("Expo export failed")
				return fmt.Errorf("expo export failed: %w", err)
			}
			spinner.Success("Exported android bundle")
		} else {
			ui.Info.Println("Skipped expo export")
		}

		err := publish("android", &client.CreateUpdateRequest{
			ProjectID:         projectCfg.ProjectID,
			RolloutPercentage: rolloutFlag,
			Channel:           projectCfg.Channel,
			RuntimeVersion:    appJson.Expo.RuntimeVersion,
			Platform:          "android",
		})

		if err != nil {
			return err
		}
	}
	if platformFlag == "all" || platformFlag == "ios" {
		if !skipExport {
			spinner, _ := ui.StartSpinner("Running expo export for ios...")
			err = runExpoExport(projectRoot, "ios")
			if err != nil {
				spinner.Fail("Expo export failed")
				return fmt.Errorf("expo export failed: %w", err)
			}
			spinner.Success("Exported ios bundle")
		} else {
			ui.Info.Println("Skipped expo export")
		}
		err := publish("ios", &client.CreateUpdateRequest{
			ProjectID:         projectCfg.ProjectID,
			RolloutPercentage: rolloutFlag,
			Channel:           projectCfg.Channel,
			RuntimeVersion:    appJson.Expo.RuntimeVersion,
			Platform:          "ios",
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func readAppJson(projectRoot string) (*AppJson, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, "app.json"))
	if err != nil {
		return nil, err
	}

	var appJson AppJson
	if err := json.Unmarshal(data, &appJson); err != nil {
		return nil, err
	}
	if appJson.Expo.RuntimeVersion == "" {
		return nil, fmt.Errorf("no runtimeVersion in app.json")
	}
	return &appJson, nil
}

func runExpoExport(projectRoot string, platform string) error {
	cmd := exec.Command("npx", "expo", "export", "--platform", platform)
	cmd.Dir = projectRoot
	// We no longer attach os.Stdout so it doesn't disrupt our spinner.
	// Only returning error is fine. If needed later, we can capture output.
	return cmd.Run()
}

func zipDistFolder(projectRoot, platform string) (string, error) {
	distDir := filepath.Join(projectRoot, "dist")

	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		return "", fmt.Errorf("dist folder not found. Run without --skip-export once")
	}

	zipPath := filepath.Join(os.TempDir(),
		fmt.Sprintf("%s-bundle-%d.zip", platform, os.Getpid()),
	)

	otherPlatform := "ios"
	if platform == "ios" {
		otherPlatform = "android"
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zipFile)

	err = filepath.Walk(distDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			target := filepath.Join("_expo", "static", "js", otherPlatform)
			if strings.Contains(path, target) {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(distDir, path)

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	zipWriter.Close()
	zipFile.Close()

	if err != nil {
		os.Remove(zipPath)
		return "", err
	}
	return zipPath, nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
