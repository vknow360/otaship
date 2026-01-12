package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const Version = "1.0.0"

// Config represents the otaship.json configuration file.
type Config struct {
	Server  string `json:"server"`
	APIKey  string `json:"api"`
	Channel string `json:"channel"`
}

func main() {
	if len(os.Args) < 2 {
		runPublish()
		return
	}

	switch os.Args[1] {
	case "init":
		runInit()
	case "publish":
		runPublish()
	case "install":
		runInstall()
	case "list", "ls":
		runList()
	case "status":
		runStatus()
	case "whoami":
		runWhoami()
	case "rollback":
		runRollback()
	case "delete", "rm":
		runDelete()
	case "doctor":
		runDoctor()
	case "upgrade":
		runUpgrade()
	case "version", "-v", "--version":
		fmt.Printf("otaship version %s\n", Version)
	case "help", "-h", "--help":
		printHelp()
	default:
		// Check if it's a flag for publish
		if strings.HasPrefix(os.Args[1], "-") {
			runPublish()
		} else {
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Run 'otaship help' for usage.")
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Println(`OTAShip CLI - Publish OTA updates for Expo apps

Usage:
  otaship [command] [flags]

Commands:
  init        Initialize otaship.json configuration
  publish     Publish an update (default command)
  list        List recent updates for current project
  rollback    Rollback to a previous update
  delete      Delete an update by ID
  status      Check server connectivity and config
  whoami      Show current API key info
  doctor      Diagnose common issues
  upgrade     Update CLI to latest version
  install     Install CLI globally to system PATH
  version     Show version information
  help        Show this help message

Publish Flags:
  --project      Path to Expo project (default: current directory)
  --server       OTAShip server URL (overrides config)
  --api          API key for authentication (overrides config)
  --channel      Update channel (default: production)
  --rollout      Rollout percentage 0-100 (default: 100)
  --skip-export  Skip expo export step

List Flags:
  --limit        Number of updates to show (default: 10)

Examples:
  otaship init                    # Create configuration file
  otaship                         # Publish using config file
  otaship list                    # View recent updates
  otaship rollback <update-id>    # Rollback to specific update
  otaship delete <update-id>      # Delete an update
  otaship doctor                  # Check for issues
  otaship status                  # Check server health`)
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("OTAShip Configuration")
	fmt.Println("=====================")
	fmt.Println("This will create an otaship.json file in the current directory.")
	fmt.Println()

	// Server URL
	fmt.Print("Server URL (default: http://localhost:8080): ")
	server, _ := reader.ReadString('\n')
	server = strings.TrimSpace(server)
	if server == "" {
		server = "http://localhost:8080"
	}

	// API Key
	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		fmt.Println("Error: API Key is required")
		os.Exit(1)
	}

	// Channel
	fmt.Print("Default channel (default: production): ")
	channel, _ := reader.ReadString('\n')
	channel = strings.TrimSpace(channel)
	if channel == "" {
		channel = "production"
	}

	config := Config{
		Server:  server,
		APIKey:  apiKey,
		Channel: channel,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error: Failed to create config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("otaship.json", data, 0644); err != nil {
		fmt.Printf("Error: Failed to write otaship.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Configuration saved to otaship.json")
	fmt.Println("You can now run 'otaship' to publish updates.")
}

func runInstall() {
	fmt.Println("OTAShip CLI Installer")
	fmt.Println("=====================")
	fmt.Println()

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error: Could not determine executable path: %v\n", err)
		os.Exit(1)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		fmt.Printf("Error: Could not resolve executable path: %v\n", err)
		os.Exit(1)
	}

	// Determine install directory based on OS
	var installDir string
	var binaryName string

	switch runtime.GOOS {
	case "windows":
		homeDir, _ := os.UserHomeDir()
		installDir = filepath.Join(homeDir, "bin")
		binaryName = "otaship.exe"
	case "darwin", "linux":
		homeDir, _ := os.UserHomeDir()
		installDir = filepath.Join(homeDir, ".local", "bin")
		binaryName = "otaship"
	default:
		fmt.Printf("Error: Unsupported operating system: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	// Create install directory if it doesn't exist
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Printf("Error: Could not create directory %s: %v\n", installDir, err)
		os.Exit(1)
	}

	destPath := filepath.Join(installDir, binaryName)

	// Check if already installed at destination
	if execPath == destPath {
		fmt.Println("OTAShip CLI is already installed globally.")
		fmt.Printf("Location: %s\n", destPath)
		return
	}

	// Copy the binary
	fmt.Printf("Installing to: %s\n", destPath)

	srcFile, err := os.Open(execPath)
	if err != nil {
		fmt.Printf("Error: Could not open source binary: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Error: Could not create destination binary: %v\n", err)
		os.Exit(1)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		fmt.Printf("Error: Could not copy binary: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Installation complete!")

	// Add to PATH
	if runtime.GOOS == "windows" {
		if addToWindowsPath(installDir) {
			fmt.Println()
			fmt.Println("Added to PATH successfully!")
			fmt.Println("Please restart your terminal to use 'otaship' from anywhere.")
		} else {
			fmt.Println()
			fmt.Println("Could not add to PATH automatically.")
			fmt.Println("Please add manually: " + installDir)
		}
	} else {
		shell := os.Getenv("SHELL")
		rcFile := "~/.bashrc"
		if strings.Contains(shell, "zsh") {
			rcFile = "~/.zshrc"
		}

		fmt.Println()
		fmt.Println("To use 'otaship' from anywhere, add to your PATH:")
		fmt.Printf("  echo 'export PATH=\"%s:$PATH\"' >> %s\n", installDir, rcFile)
		fmt.Printf("  source %s\n", rcFile)
	}
}

// addToWindowsPath adds a directory to the user's PATH environment variable on Windows.
func addToWindowsPath(dir string) bool {
	// Get current user PATH
	cmd := exec.Command("powershell", "-Command",
		"[Environment]::GetEnvironmentVariable('Path', 'User')")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	currentPath := strings.TrimSpace(string(output))

	// Check if already in PATH
	paths := strings.Split(currentPath, ";")
	for _, p := range paths {
		if strings.EqualFold(strings.TrimSpace(p), dir) {
			fmt.Println("Directory already in PATH.")
			return true
		}
	}

	// Add to PATH
	var newPath string
	if currentPath == "" {
		newPath = dir
	} else {
		newPath = currentPath + ";" + dir
	}

	cmd = exec.Command("powershell", "-Command",
		fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', 'User')", newPath))
	err = cmd.Run()
	return err == nil
}

func runPublish() {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)

	projectPath := fs.String("project", ".", "Path to Expo project")
	channel := fs.String("channel", "", "Update channel")
	rollout := fs.Int("rollout", 100, "Rollout percentage (0-100)")
	serverURL := fs.String("server", "", "OTAShip server URL")
	apiKey := fs.String("api", "", "API Key")
	skipExport := fs.Bool("skip-export", false, "Skip expo export")

	// Parse args, skipping "publish" if present
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "publish" {
		args = args[1:]
	}
	fs.Parse(args)

	// Load config
	config := loadConfig(*projectPath)

	// Override with flags
	if *serverURL != "" {
		config.Server = *serverURL
	}
	if *apiKey != "" {
		config.APIKey = *apiKey
	}
	if *channel != "" {
		config.Channel = *channel
	}
	if config.Channel == "" {
		config.Channel = "production"
	}

	// Validate
	if config.Server == "" {
		fmt.Println("Error: Server URL is required")
		fmt.Println("Run 'otaship init' to create a configuration file.")
		os.Exit(1)
	}
	if config.APIKey == "" {
		fmt.Println("Error: API Key is required")
		fmt.Println("Run 'otaship init' to create a configuration file.")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("OTAShip Publisher v%s\n", Version)
	fmt.Println("========================")
	fmt.Println()

	// Step 1: Export
	if !*skipExport {
		fmt.Println("[1/3] Exporting Expo bundle...")
		if err := runExpoExport(*projectPath); err != nil {
			fmt.Printf("Error: Export failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("      Export complete")
	} else {
		fmt.Println("[1/3] Skipping export (--skip-export)")
	}

	// Step 2: Read config
	fmt.Println("[2/3] Reading project configuration...")
	projectSlug, runtimeVersion, err := readAppConfig(*projectPath)
	if err != nil {
		fmt.Printf("Error: Failed to read app.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("      Project: %s (runtime: %s)\n", projectSlug, runtimeVersion)

	// Step 3: Publish
	fmt.Println("[3/3] Publishing to server...")
	distPath := filepath.Join(*projectPath, "dist")
	distPath, _ = filepath.Abs(distPath)
	updateID := generateUUID()

	if err := publishUpdate(config.Server, config.APIKey, projectSlug, updateID, runtimeVersion, config.Channel, distPath, *rollout); err != nil {
		fmt.Printf("Error: Publish failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Update published successfully!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("  Project:  %s\n", projectSlug)
	fmt.Printf("  Runtime:  %s\n", runtimeVersion)
	fmt.Printf("  Channel:  %s\n", config.Channel)
	fmt.Printf("  Rollout:  %d%%\n", *rollout)
	fmt.Printf("  Server:   %s\n", config.Server)
	fmt.Println()
}

// runList lists recent updates from the server
func runList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	limit := fs.Int("limit", 10, "Number of updates to show")
	projectPath := fs.String("project", ".", "Path to Expo project")

	args := os.Args[2:]
	fs.Parse(args)

	config := loadConfig(*projectPath)
	if config.Server == "" {
		fmt.Println("Error: Server URL is required")
		fmt.Println("Run 'otaship init' to create a configuration file.")
		os.Exit(1)
	}
	if config.APIKey == "" {
		fmt.Println("Error: API Key is required")
		fmt.Println("Run 'otaship init' to create a configuration file.")
		os.Exit(1)
	}

	// Fetch updates
	url := fmt.Sprintf("%s/api/admin/updates?limit=%d", strings.TrimRight(config.Server, "/"), *limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: Could not connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: Server returned %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var result struct {
		Updates []struct {
			ID             string    `json:"id"`
			UpdateID       string    `json:"updateId"`
			ProjectSlug    string    `json:"projectSlug"`
			RuntimeVersion string    `json:"runtimeVersion"`
			Channel        string    `json:"channel"`
			IsActive       bool      `json:"isActive"`
			Downloads      int       `json:"downloads"`
			CreatedAt      time.Time `json:"createdAt"`
		} `json:"updates"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Recent Updates (showing %d of %d)\n", len(result.Updates), result.Total)
	fmt.Println("=" + strings.Repeat("=", 75))
	fmt.Println()

	if len(result.Updates) == 0 {
		fmt.Println("No updates found.")
		return
	}

	// Table header
	fmt.Printf("%-20s %-12s %-12s %-8s %s\n", "PROJECT", "RUNTIME", "CHANNEL", "ACTIVE", "CREATED")
	fmt.Println(strings.Repeat("-", 76))

	for _, u := range result.Updates {
		active := " "
		if u.IsActive {
			active = "*"
		}
		fmt.Printf("%-20s %-12s %-12s %-8s %s\n",
			truncate(u.ProjectSlug, 20),
			u.RuntimeVersion,
			u.Channel,
			active,
			u.CreatedAt.Format("2006-01-02"),
		)
	}
	fmt.Println()
}

// runStatus checks server connectivity and shows config
func runStatus() {
	projectPath := "."
	if len(os.Args) > 2 {
		projectPath = os.Args[2]
	}

	config := loadConfig(projectPath)

	fmt.Println()
	fmt.Println("OTAShip Status")
	fmt.Println("==============")
	fmt.Println()

	// Show config
	fmt.Println("Configuration:")
	if config.Server != "" {
		fmt.Printf("  Server:  %s\n", config.Server)
	} else {
		fmt.Println("  Server:  (not configured)")
	}
	if config.APIKey != "" {
		fmt.Printf("  API Key: %s...%s\n", config.APIKey[:4], config.APIKey[len(config.APIKey)-4:])
	} else {
		fmt.Println("  API Key: (not configured)")
	}
	fmt.Printf("  Channel: %s\n", config.Channel)
	fmt.Println()

	// Check server health
	if config.Server == "" {
		fmt.Println("Server:  Not configured. Run 'otaship init'.")
		return
	}

	fmt.Print("Checking server... ")
	url := fmt.Sprintf("%s/api/health", strings.TrimRight(config.Server, "/"))
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("OFFLINE (%v)\n", err)
		return
	}
	defer resp.Body.Close()

	var health struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	json.NewDecoder(resp.Body).Decode(&health)

	if health.Status == "ok" {
		fmt.Printf("ONLINE (v%s)\n", health.Version)
	} else {
		fmt.Printf("DEGRADED (%s)\n", health.Status)
	}
	fmt.Println()
}

// runWhoami shows current API key info
func runWhoami() {
	projectPath := "."
	if len(os.Args) > 2 {
		projectPath = os.Args[2]
	}

	config := loadConfig(projectPath)

	fmt.Println()
	if config.APIKey == "" {
		fmt.Println("Not configured. Run 'otaship init' to set up.")
		return
	}

	fmt.Println("Current Configuration")
	fmt.Println("=====================")
	fmt.Printf("Server:  %s\n", config.Server)
	fmt.Printf("API Key: %s...%s\n", config.APIKey[:min(4, len(config.APIKey))], config.APIKey[max(0, len(config.APIKey)-4):])
	fmt.Printf("Channel: %s\n", config.Channel)
	fmt.Println()
}

// runRollback rolls back to a previous update
func runRollback() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Update ID is required")
		fmt.Println("Usage: otaship rollback <update-id>")
		os.Exit(1)
	}

	updateID := os.Args[2]
	config := loadConfig(".")

	if config.Server == "" || config.APIKey == "" {
		fmt.Println("Error: Configuration required. Run 'otaship init' first.")
		os.Exit(1)
	}

	fmt.Printf("Rolling back to update: %s\n", updateID)
	fmt.Print("Confirming rollback... ")

	url := fmt.Sprintf("%s/api/admin/updates/%s/rollback", strings.TrimRight(config.Server, "/"), updateID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED\nError: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("FAILED\nServer returned %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	fmt.Println("SUCCESS")
	fmt.Println()
	fmt.Printf("Update %s is now the active version.\n", updateID)
	fmt.Println()
}

// runDelete deletes an update by ID
func runDelete() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Update ID is required")
		fmt.Println("Usage: otaship delete <update-id>")
		os.Exit(1)
	}

	updateID := os.Args[2]
	config := loadConfig(".")

	if config.Server == "" || config.APIKey == "" {
		fmt.Println("Error: Configuration required. Run 'otaship init' first.")
		os.Exit(1)
	}

	// Confirm deletion
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Delete update %s? This cannot be undone. [y/N]: ", updateID)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	fmt.Print("Deleting update... ")

	url := fmt.Sprintf("%s/api/admin/updates/%s", strings.TrimRight(config.Server, "/"), updateID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED\nError: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("FAILED\nServer returned %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	fmt.Println("SUCCESS")
	fmt.Printf("Update %s has been deleted.\n", updateID)
}

// runDoctor diagnoses common issues
func runDoctor() {
	fmt.Println()
	fmt.Println("OTAShip Doctor")
	fmt.Println("==============")
	fmt.Println()

	allGood := true

	// Check 1: Configuration file
	fmt.Print("[1/5] Checking configuration... ")
	config := loadConfig(".")
	if config.Server != "" && config.APIKey != "" {
		fmt.Println("OK")
	} else if config.Server == "" && config.APIKey == "" {
		fmt.Println("NOT CONFIGURED")
		fmt.Println("      Run 'otaship init' to create configuration.")
		allGood = false
	} else {
		fmt.Println("INCOMPLETE")
		if config.Server == "" {
			fmt.Println("      Missing: server URL")
		}
		if config.APIKey == "" {
			fmt.Println("      Missing: API key")
		}
		allGood = false
	}

	// Check 2: Server connectivity
	fmt.Print("[2/5] Checking server... ")
	if config.Server != "" {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(fmt.Sprintf("%s/api/health", strings.TrimRight(config.Server, "/")))
		if err != nil {
			fmt.Println("OFFLINE")
			fmt.Printf("      Could not connect: %v\n", err)
			allGood = false
		} else {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				fmt.Println("OK")
			} else {
				fmt.Printf("ERROR (status %d)\n", resp.StatusCode)
				allGood = false
			}
		}
	} else {
		fmt.Println("SKIPPED (no server configured)")
	}

	// Check 3: Expo CLI
	fmt.Print("[3/5] Checking Expo CLI... ")
	cmd := exec.Command("npx", "expo", "--version")
	if output, err := cmd.Output(); err != nil {
		fmt.Println("NOT FOUND")
		fmt.Println("      Install with: npm install -g expo-cli")
		allGood = false
	} else {
		version := strings.TrimSpace(string(output))
		fmt.Printf("OK (v%s)\n", version)
	}

	// Check 4: app.json
	fmt.Print("[4/5] Checking app.json... ")
	if _, err := os.Stat("app.json"); os.IsNotExist(err) {
		fmt.Println("NOT FOUND")
		fmt.Println("      This doesn't appear to be an Expo project directory.")
		allGood = false
	} else {
		// Check for required fields
		slug, runtime, err := readAppConfig(".")
		if err != nil {
			fmt.Printf("ERROR (%v)\n", err)
			allGood = false
		} else if slug == "" {
			fmt.Println("MISSING SLUG")
			fmt.Println("      Add 'expo.slug' to your app.json")
			allGood = false
		} else {
			fmt.Printf("OK (slug: %s, runtime: %s)\n", slug, runtime)
		}
	}

	// Check 5: Node modules
	fmt.Print("[5/5] Checking node_modules... ")
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		fmt.Println("NOT FOUND")
		fmt.Println("      Run 'npm install' first.")
		allGood = false
	} else {
		fmt.Println("OK")
	}

	fmt.Println()
	if allGood {
		fmt.Println("All checks passed! You're ready to publish.")
	} else {
		fmt.Println("Some issues were found. Please fix them before publishing.")
	}
	fmt.Println()
}

// GitHub repository for releases
const GitHubRepo = "vknow360/otaship"

// runUpgrade updates the CLI to the latest version from GitHub releases
func runUpgrade() {
	fmt.Println()
	fmt.Println("OTAShip Upgrade")
	fmt.Println("===============")
	fmt.Println()
	fmt.Printf("Current version: %s\n", Version)
	fmt.Print("Checking for updates... ")

	// Fetch latest release from GitHub API
	client := &http.Client{Timeout: 30 * time.Second}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "otaship-cli/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED\nError: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Println("NO RELEASES")
		fmt.Println("No releases found. The repository may not have any releases yet.")
		return
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("FAILED\nGitHub API returned %d: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("FAILED\nError parsing response: %v\n", err)
		os.Exit(1)
	}

	// Extract version (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	fmt.Printf("FOUND v%s\n", latestVersion)

	// Compare versions
	if latestVersion == Version {
		fmt.Println()
		fmt.Println("You're already on the latest version!")
		return
	}

	fmt.Printf("New version available: %s -> %s\n", Version, latestVersion)
	fmt.Println()

	// Determine which asset to download based on OS and architecture
	var assetName string
	switch runtime.GOOS {
	case "windows":
		assetName = "otaship-windows-amd64.exe"
		if runtime.GOARCH == "arm64" {
			assetName = "otaship-windows-arm64.exe"
		}
	case "darwin":
		assetName = "otaship-darwin-amd64"
		if runtime.GOARCH == "arm64" {
			assetName = "otaship-darwin-arm64"
		}
	case "linux":
		assetName = "otaship-linux-amd64"
		if runtime.GOARCH == "arm64" {
			assetName = "otaship-linux-arm64"
		}
	default:
		fmt.Printf("Error: Unsupported OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	// Find matching asset
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		fmt.Printf("Error: No binary found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Available assets:")
		for _, asset := range release.Assets {
			fmt.Printf("  - %s\n", asset.Name)
		}
		os.Exit(1)
	}

	// Download the new binary
	fmt.Printf("Downloading %s... ", assetName)
	resp, err = client.Get(downloadURL)
	if err != nil {
		fmt.Printf("FAILED\nError: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("FAILED\nHTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("FAILED\nCould not determine executable path: %v\n", err)
		os.Exit(1)
	}
	execPath, _ = filepath.EvalSymlinks(execPath)

	// Create temp file for download (with .exe extension on Windows)
	tmpExt := ""
	if runtime.GOOS == "windows" {
		tmpExt = ".exe"
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "otaship-upgrade-*"+tmpExt)
	if err != nil {
		fmt.Printf("FAILED\nCould not create temp file: %v\n", err)
		os.Exit(1)
	}
	tmpPath := tmpFile.Name()

	// Download to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		fmt.Printf("FAILED\nDownload error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	// Replace the current executable
	fmt.Print("Installing update... ")

	// On Windows, we can't delete/overwrite a running executable
	// But we CAN rename it, then rename the new file into place
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		os.Remove(oldPath) // Remove any existing .old file

		// Rename running exe to .old
		if err := os.Rename(execPath, oldPath); err != nil {
			os.Remove(tmpPath)
			fmt.Printf("FAILED\nCould not rename old binary: %v\n", err)
			os.Exit(1)
		}

		// Rename temp file to target
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try to restore old file
			os.Rename(oldPath, execPath)
			os.Remove(tmpPath)
			fmt.Printf("FAILED\nCould not install new binary: %v\n", err)
			os.Exit(1)
		}

		// Schedule cleanup of .old file (can't delete while running)
		// It will be cleaned up on next upgrade
	} else {
		// On Unix, we can overwrite with rename
		os.Chmod(tmpPath, 0755)
		if err := os.Rename(tmpPath, execPath); err != nil {
			os.Remove(tmpPath)
			fmt.Printf("FAILED\n%v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("OK")
	fmt.Println()
	fmt.Printf("Successfully upgraded to v%s!\n", latestVersion)
	fmt.Println("Run 'otaship version' to verify.")
	fmt.Println()
}

// copyFile copies a file from src to dst
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

// truncate shortens a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func loadConfig(projectPath string) Config {
	configPath := filepath.Join(projectPath, "otaship.json")
	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Warning: Failed to parse otaship.json: %v\n", err)
	}

	return config
}

func runExpoExport(projectPath string) error {
	cmd := exec.Command("npx", "expo", "export")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func readAppConfig(projectPath string) (string, string, error) {
	appJsonPath := filepath.Join(projectPath, "app.json")
	data, err := os.ReadFile(appJsonPath)
	if err != nil {
		return "", "", err
	}

	var config struct {
		Expo struct {
			Slug           string `json:"slug"`
			RuntimeVersion string `json:"runtimeVersion"`
		} `json:"expo"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return "", "", err
	}

	if config.Expo.Slug == "" {
		return "", "", fmt.Errorf("expo.slug is required in app.json")
	}

	runtimeVersion := config.Expo.RuntimeVersion
	if runtimeVersion == "" {
		runtimeVersion = "1"
	}

	return config.Expo.Slug, runtimeVersion, nil
}

// zipDirectory zips the contents of the specified directory into a target file.
// If includeFiles is non-nil, only files in the map are included.
// Keys in includeFiles should be slash-separated relative paths (e.g. "dist/metadata.json").
// overrides allows substituting a file content with another local file path.
func zipDirectory(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(sourceDir)
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path for filtering
		relPath, _ := filepath.Rel(sourceDir, path)
		relPath = filepath.ToSlash(relPath) // Standardize to forward slash

		// Prepare header
		var header *zip.FileHeader

		header, err = zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			parts := strings.Split(relPath, "/")
			if relPath == "." {
				parts = []string{}
			}
			joinParts := append([]string{baseDir}, parts...)
			header.Name = filepath.ToSlash(filepath.Join(joinParts...))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var file *os.File
		file, err = os.Open(path)

		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func publishUpdate(serverURL, token, projectSlug, updateID, runtimeVersion, channel, distPath string, rollout int) error {
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return fmt.Errorf("dist directory not found at %s", distPath)
	}

	// 1. Create Zip Bundle
	fmt.Println("      Compressing bundle...")
	tmpFile, err := os.CreateTemp("", "otaship-bundle-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Cleanup
	tmpFile.Close()                 // Close so zipDirectory can open it

	if err := zipDirectory(distPath, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to zip bundle: %v", err)
	}

	// 2. Prepare Multipart Upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields
	_ = writer.WriteField("projectSlug", projectSlug)
	_ = writer.WriteField("updateId", updateID)
	_ = writer.WriteField("runtimeVersion", runtimeVersion)
	_ = writer.WriteField("channel", channel)
	_ = writer.WriteField("rolloutPercentage", fmt.Sprintf("%d", rollout))
	_ = writer.WriteField("platform", "android") // Should optionally come from config, defaulting to all/android

	// Add file
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("bundle", "bundle.zip")
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// 3. Send Request
	startTime := time.Now()
	url := fmt.Sprintf("%s/api/admin/updates", strings.TrimRight(serverURL, "/"))

	// Create progress reader
	contentLength := int64(body.Len())
	progressReader := &ProgressReader{
		Reader: body,
		Total:  contentLength,
	}

	req, err := http.NewRequest("POST", url, progressReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Minute} // Large timeout for upload
	req.ContentLength = contentLength                 // Set content length explicitly

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Clear progress line
	fmt.Printf("\r      Uploading... 100%% (%s / %s) - Done in %s\n", formatBytes(contentLength), formatBytes(contentLength), time.Since(startTime).Round(time.Second))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ProgressReader wraps an io.Reader to track read progress
type ProgressReader struct {
	io.Reader
	Total   int64
	Current int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	val := float64(pr.Current) / float64(pr.Total) * 100
	fmt.Printf("\r      Uploading... %.1f%% (%s / %s)   ", val, formatBytes(pr.Current), formatBytes(pr.Total))

	return n, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
