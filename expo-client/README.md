# OTAShip Expo Client

Example Expo application demonstrating OTA update integration with OTAShip.

Use this as a reference for configuring your own Expo apps to receive OTA updates.

## Setup

### Prerequisites

- Node.js 18+
- Expo CLI
- OTAShip backend running

### Installation

```bash
# Install dependencies
npm install

# Start development server
npx expo start
```

### Building for Device

```bash
# Android
npx expo run:android

# iOS
npx expo run:ios
```

## OTA Configuration

The key configuration is in `app.json`:

```json
{
  "expo": {
    "slug": "expo-client",
    "runtimeVersion": "2",
    "updates": {
      "url": "https://your-server.com/api/expo-client/manifest",
      "enabled": true,
      "checkAutomatically": "ON_LOAD"
    }
  }
}
```

### Configuration Options

| Field                        | Description                                   |
| ---------------------------- | --------------------------------------------- |
| `slug`                       | Unique project identifier (must match server) |
| `runtimeVersion`             | Version string for update compatibility       |
| `updates.url`                | Your OTAShip manifest endpoint                |
| `updates.enabled`            | Enable/disable OTA updates                    |
| `updates.checkAutomatically` | When to check for updates                     |

### Runtime Version

The `runtimeVersion` determines update compatibility:

- Updates are only delivered to clients with matching runtime versions
- Increment when making native code changes
- Keep same for JavaScript-only updates

## Checking for Updates

The example app demonstrates manual update checking:

```javascript
import * as Updates from "expo-updates";

async function checkForUpdates() {
  const update = await Updates.checkForUpdateAsync();
  if (update.isAvailable) {
    await Updates.fetchUpdateAsync();
    await Updates.reloadAsync();
  }
}
```

## File Structure

```
expo-client/
├── App.js              # Main app with update UI
├── app.json            # Expo configuration
├── assets/             # App icons and images
└── package.json        # Dependencies
```

## License

MIT
