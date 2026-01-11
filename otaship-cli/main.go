package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

Examples:
  otaship init                    # Create configuration file
  otaship                         # Publish using config file
  otaship --channel staging       # Publish to staging channel
  otaship --rollout 50            # Publish to 50% of users
  otaship install                 # Install CLI globally`)
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
	fmt.Println()

	// Print PATH instructions
	if runtime.GOOS == "windows" {
		fmt.Println("To use 'otaship' from anywhere, add the install directory to your PATH:")
		fmt.Println()
		fmt.Println("  Option 1: Run this command in PowerShell (as Administrator):")
		fmt.Printf("    [Environment]::SetEnvironmentVariable(\"Path\", $env:Path + \";%s\", \"User\")\n", installDir)
		fmt.Println()
		fmt.Println("  Option 2: Manually add to PATH:")
		fmt.Println("    1. Open System Properties > Environment Variables")
		fmt.Println("    2. Under 'User variables', edit 'Path'")
		fmt.Printf("    3. Add: %s\n", installDir)
		fmt.Println()
		fmt.Println("After updating PATH, restart your terminal.")
	} else {
		shell := os.Getenv("SHELL")
		rcFile := "~/.bashrc"
		if strings.Contains(shell, "zsh") {
			rcFile = "~/.zshrc"
		}

		fmt.Println("To use 'otaship' from anywhere, add the install directory to your PATH:")
		fmt.Println()
		fmt.Printf("  echo 'export PATH=\"%s:$PATH\"' >> %s\n", installDir, rcFile)
		fmt.Printf("  source %s\n", rcFile)
	}
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

func publishUpdate(serverURL, token, projectSlug, updateID, runtimeVersion, channel, distPath string, rollout int) error {
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return fmt.Errorf("dist directory not found at %s", distPath)
	}

	reqBody := map[string]interface{}{
		"projectSlug":       projectSlug,
		"updateId":          updateID,
		"runtimeVersion":    runtimeVersion,
		"channel":           channel,
		"rolloutPercentage": rollout,
		"bundlePath":        distPath,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/admin/updates", strings.TrimRight(serverURL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
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
