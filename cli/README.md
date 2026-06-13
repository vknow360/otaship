# OTAShip — CLI

The command-line tool for publishing OTA updates to your OTAShip backend. Works locally and in CI/CD pipelines.

→ [Back to main README](../README.md)

## Tech Stack

| | |
|---|---|
| **Language** | Go 1.25+ |
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) |
| **Bundling** | Runs `npx expo export` and zips the output |

## How It Works

The CLI bridges your local Expo project with the OTAShip backend:

1. Runs `npx expo export` to generate the JS bundle and assets
2. Zips the `dist/` output into a single archive
3. Uploads it to your OTAShip backend using the project's API key
4. The backend processes the bundle, stores assets, and makes the update available

## Installation

**From source:**

```bash
cd cli
go install ./cmd/otaship
```

**Pre-built binaries:** Download from the [Releases](https://github.com/vknow360/otaship/releases) page.

## Usage

### 1. Login to your server

```bash
otaship login
```

Prompts for your OTAShip server URL (e.g., `https://api.yourdomain.com`).

### 2. Initialize a project

```bash
cd your-expo-app
otaship init
```

Prompts for your project's API key and creates an `otaship.json` config file.

### 3. Publish an update

```bash
otaship publish --message "Fixed crash on login screen"
```

This exports, bundles, and uploads in one step.

### Command Reference

#### `otaship publish`

Bundles the Expo project and publishes an update to your server.

| Flag | Default | Description |
|------|---------|-------------|
| `--platform` | `all` | Target platform: `android`, `ios`, or `all` |
| `--channel` | from config | Override the release channel |
| `--rollout` | `100` | Percentage of users to receive this update (0–100) |
| `--message` | | Changelog or description for this update |
| `--skip-export` | `false` | Skip `npx expo export` (use existing `dist/`) |
| `--dry-run` | `false` | Bundle locally without uploading |
| `-y, --yes` | `false` | Skip confirmation prompts (useful for CI/CD) |

#### `otaship rollback <update-id>`

Republishes a previous update to the active channel, making it the current update again.

#### `otaship reset`

Instructs all clients to revert to the embedded app binary — effectively a factory reset.

### CI/CD Example

```yaml
# GitHub Actions
- name: Publish OTA update
  run: |
    otaship login
    otaship publish --platform android --channel production --yes
  env:
    OTASHIP_SERVER_URL: ${{ secrets.OTASHIP_URL }}
```

## Project Structure

```
cli/
├── cmd/otaship/         # Entry point
└── internal/
    ├── client/          # HTTP client for backend API
    ├── commands/        # Cobra command definitions
    ├── config/          # otaship.json reading/writing
    ├── ui/              # Terminal output formatting
    └── utils/           # Shared helpers (zip, hash, etc.)
```