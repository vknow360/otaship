# OTAShip CLI

Command-line tool for publishing OTA updates to your OTAShip server.

## Installation

### Option 1: Build and Install Globally (Recommended)

```bash
# Clone and build
cd otaship-cli
go build -o otaship.exe .   # Windows
go build -o otaship .       # macOS/Linux

# Run the installer
./otaship install
```

The `install` command copies the binary to your user directory and provides instructions to add it to your system PATH.

### Option 2: Go Install

```bash
go install github.com/vknow360/otaship/otaship-cli@latest
```

### Option 3: Manual Installation

1. Build the binary: `go build -o otaship .`
2. Move to a directory in your PATH (e.g., `/usr/local/bin` or `C:\Windows\System32`)

## Quick Start

```bash
# Navigate to your Expo project
cd my-expo-app

# Create configuration file (one-time setup)
otaship init

# Publish an update
otaship
```

## Commands

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `otaship init`    | Create `otaship.json` configuration file |
| `otaship`         | Publish an update (default command)      |
| `otaship publish` | Publish an update (explicit)             |
| `otaship install` | Install CLI globally to system PATH      |
| `otaship version` | Show version information                 |
| `otaship help`    | Show help message                        |

## Configuration

The CLI uses `otaship.json` in your project directory:

```json
{
  "server": "https://your-otaship-server.com",
  "api": "your-api-key",
  "channel": "production"
}
```

| Field     | Description                |
| --------- | -------------------------- |
| `server`  | OTAShip server URL         |
| `api`     | API key for authentication |
| `channel` | Default release channel    |

## Publish Options

| Flag            | Default      | Description                |
| --------------- | ------------ | -------------------------- |
| `--project`     | `.`          | Path to Expo project       |
| `--server`      | from config  | OTAShip server URL         |
| `--api`         | from config  | API key for authentication |
| `--channel`     | `production` | Update channel             |
| `--rollout`     | `100`        | Rollout percentage (0-100) |
| `--skip-export` | `false`      | Skip `expo export` step    |

## Examples

```bash
# Publish with default settings (uses otaship.json)
otaship

# Publish to staging channel
otaship --channel staging

# Publish with 50% rollout
otaship --rollout 50

# Publish to a different server
otaship --server https://staging.otaship.com --api staging-key

# Skip export if already done
otaship --skip-export

# Publish a different project
otaship --project ../my-other-app
```

## Workflow

1. **Setup** (one-time): Run `otaship init` in your Expo project
2. **Develop**: Make changes to your app
3. **Publish**: Run `otaship` to push updates

The CLI will:

1. Run `expo export` to build your bundle
2. Read project info from `app.json`
3. Upload the bundle to your OTAShip server

## Troubleshooting

### "Server URL is required"

Run `otaship init` to create a configuration file, or use `--server` flag.

### "API Key is required"

Run `otaship init` to set your API key, or use `--api` flag.

### "dist directory not found"

Make sure `expo export` completed successfully. Check for errors in the export output.

### "expo.slug is required in app.json"

Ensure your `app.json` has the `expo.slug` field set.

## Runtime Version

The CLI reads `runtimeVersion` from your `app.json`. Two formats are supported:

### String Format

```json
{
  "expo": {
    "runtimeVersion": "1.0.0"
  }
}
```

### Policy Format

When using a policy object, the CLI will resolve the runtime version based on the policy:

```json
{
  "expo": {
    "version": "1.1.0",
    "runtimeVersion": {
      "policy": "appVersion"
    }
  }
}
```

| Policy       | Description                                      |
| ------------ | ------------------------------------------------ |
| `appVersion` | Uses the `expo.version` field as runtime version |

If no `runtimeVersion` is specified, it defaults to `"1"`.

## License

MIT
