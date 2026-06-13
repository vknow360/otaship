# OTAShip — Expo Client Example

A reference React Native app that demonstrates how to receive OTA updates from a self-hosted OTAShip backend.

→ [Back to main README](../README.md)

## Tech Stack

| | |
|---|---|
| **Framework** | [Expo](https://expo.dev/) (React Native) |
| **OTA Client** | [`expo-updates`](https://docs.expo.dev/versions/latest/sdk/updates/) |
| **Code Signing** | RSA certificate verification via `codeSigningCertificate` |

## What This Does

This isn't a library you install — it's a working example showing how to wire up any Expo app with OTAShip. The app:

1. Connects to your OTAShip backend on launch
2. Compares the server's latest update against its local `runtimeVersion`
3. If a new update is available (and the device falls within the rollout window), downloads and applies the new JS bundle
4. Supports code-signed manifests for integrity verification

## Setup

```bash
cd expo-client
npm install
```

### Configure `app.json`

Point `expo-updates` at your OTAShip backend:

```json
{
  "expo": {
    "runtimeVersion": "2",
    "updates": {
      "url": "https://your-server.com/api/manifest/<project-id>",
      "enabled": true,
      "checkAutomatically": "ON_LOAD"
    }
  }
}
```

### Configuration Reference

| Field | Description |
|-------|-------------|
| `runtimeVersion` | Identifies native code compatibility. The backend only serves updates matching this version. |
| `updates.url` | Your OTAShip manifest endpoint: `<server>/api/manifest/<project-id>` |
| `updates.enabled` | Set `true` to enable OTA checking |
| `updates.checkAutomatically` | `ON_LOAD` checks on every app launch. `NEVER` requires manual checks via code. |

### Code Signing (Optional)

To verify that updates come from a trusted source, add code signing config:

```json
{
  "updates": {
    "codeSigningCertificate": "./certs/certificate.pem",
    "codeSigningMetadata": {
      "keyid": "main",
      "alg": "rsa-v1_5-sha256"
    }
  }
}
```

Place your certificate in a `certs/` directory. The corresponding private key should be configured on the backend via the `EXPO_PRIVATE_KEY` environment variable.

## Running

```bash
# Development (Metro bundler — no OTA, live reload instead)
npx expo start

# Native build (required to actually test OTA updates)
npx expo run:android
# or
npx expo run:ios
```

> **Important:** OTA updates only work in native builds (`expo run:*`), not in Expo Go or the Metro dev server.

## Manual Update Check

You can also trigger update checks programmatically:

```javascript
import * as Updates from "expo-updates";

async function checkForUpdates() {
  const update = await Updates.checkForUpdateAsync();
  if (update.isAvailable) {
    await Updates.fetchUpdateAsync();
    await Updates.reloadAsync(); // Restarts the app with the new bundle
  }
}
```

## Project Structure

```
expo-client/
├── App.js              # Main entry point with update check example
├── app.json            # Expo config (OTA URL, code signing)
├── otaship.json        # OTAShip CLI config (API key, project ID)
├── certs/              # Code signing certificate
├── keys/               # Signing keys
├── assets/             # App icons and splash screen
└── package.json        # Dependencies
```
