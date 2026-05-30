# OTAShip CLI

The OTAShip CLI is a powerful command-line tool designed to help you publish, rollback, and manage Expo OTA updates on your self-hosted OTAShip server.

## Installation

### From GitHub Releases (Recommended)
Download the latest pre-compiled binary for your operating system from the [Releases page](https://github.com/vknow360/otaship/releases). Extract it and move the `otaship` binary into your system's `PATH`.

### From Source
If you have Go 1.25+ installed, you can compile and install it directly:
```bash
go install github.com/vknow360/otaship/cli/cmd/otaship@latest
```

### Upgrading
To upgrade an existing installation to the latest version, simply run:
```bash
otaship upgrade
```

## Getting Started

Before managing updates, you need to authenticate the CLI with your OTAShip backend instance.

1. **Log in to your server:**
   ```bash
   otaship login
   # You will be prompted to enter your server URL (e.g., https://api.yourdomain.com)
   ```

2. **Connect your Expo project:**
   If you are the project creator setting up a new project:
   ```bash
   otaship init
   # You will be prompted for the API Key provided by the Admin Dashboard
   ```
   
   If you are a team member joining an existing project that already has an `otaship.json` file:
   ```bash
   otaship link
   # You will be prompted for your API Key
   ```

## Managing Updates

### Publishing a New Update
To bundle and publish your current Expo project code to OTAShip:
```bash
otaship publish
```
The CLI will interactively ask you for the platform (iOS/Android/All), release channel, rollout percentage, and an optional release message.

**Non-Interactive / CI Usage:**
You can bypass prompts for use in CI/CD pipelines (like GitHub Actions):
```bash
otaship publish --platform android --channel production --rollout 25 --message "Fix login bug" --yes
```

**Skip Export:**
If you have already run `npx expo export` manually, you can skip the export step:
```bash
otaship publish --skip-export
```

### Listing Updates
View a history of all published updates for the current project:
```bash
otaship list
```

### Rolling Back
If you published a broken update, you can instantly republish a previous known-good update using its ID:
```bash
otaship rollback <update-id>
```

### Factory Reset
To force all devices on a specific platform/channel to clear their cached OTA updates and revert to the original binary built into the app stores:
```bash
otaship reset --platform android --channel production
```

### Deleting an Update
Remove an update record entirely:
```bash
otaship delete <update-id>
```

## Other Commands

- `otaship status`: Shows connection info and current project context.
- `otaship doctor`: Validates your current setup and environment.
- `otaship whoami`: Displays which project your API key is currently linked to.
- `otaship version`: Prints the current CLI version.

## Configuration Files

The CLI maintains two types of configuration files:

- **Global Config (`~/.otaship/config.json`)**: Stores your server URL and securely maps Project IDs to their respective API keys on your machine.
- **Project Config (`./otaship.json`)**: Sits alongside your `app.json`. Stores non-sensitive project metadata like the `projectId` and default `channel`. This file should be checked into version control.