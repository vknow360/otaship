# OTAShip Backend

A self-hosted OTA (Over-The-Air) update server for Expo and React Native applications. Built with Go and the Gin framework for high performance and reliability.

## Features

- **OTA Updates** — Push updates instantly without app store review
- **Analytics** — Track downloads, adoption rates, and update performance
- **Gradual Rollouts** — Control update distribution with percentage-based rollouts
- **CDN Integration** — Cloudinary support for global asset delivery
- **Multi-Channel** — Separate production, staging, and beta release channels
- **Code Signing** — Optional RSA-SHA256 manifest signing for security

## Quick Start

### Prerequisites

- Go 1.21+
- MongoDB Atlas (or local MongoDB)
- Cloudinary account (optional, for CDN)

### Local Development

```bash
# Clone and navigate to backend
cd backend

# Copy environment template
cp .env.example .env

# Configure your environment variables in .env
# Required: MONGODB_URI
# Optional: CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET

# Run the server
go run .
```

The server will start at `http://localhost:8080`.

### Deploy to Render

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

## Configuration

| Variable                | Required | Description                         | Example                        |
| ----------------------- | -------- | ----------------------------------- | ------------------------------ |
| `PORT`                  | No       | Server port                         | `8080`                         |
| `HOSTNAME`              | Yes      | Public URL of your server           | `https://otaship.onrender.com` |
| `MONGODB_URI`           | Yes      | MongoDB connection string           | `mongodb+srv://...`            |
| `CLOUDINARY_CLOUD_NAME` | No       | Cloudinary cloud name               | `your-cloud`                   |
| `CLOUDINARY_API_KEY`    | No       | Cloudinary API key                  | `123456789`                    |
| `CLOUDINARY_API_SECRET` | No       | Cloudinary API secret               | `abc123...`                    |
| `ADMIN_SECRET`          | Yes      | Secret token for admin API          | `your-secret-token`            |
| `PRIVATE_KEY_PATH`      | No       | Path to RSA private key for signing | `./keys/private.pem`           |

## API Reference

### Client Endpoints

These endpoints are called by Expo apps to check for and download updates.

| Method | Endpoint                     | Description                       |
| ------ | ---------------------------- | --------------------------------- |
| `GET`  | `/api/:projectSlug/manifest` | Get update manifest for a project |
| `GET`  | `/api/:projectSlug/assets`   | Download update assets            |
| `GET`  | `/api/health`                | Health check                      |

### Admin Endpoints

Protected endpoints for managing updates. Requires `Authorization: Bearer <ADMIN_SECRET>` header.

| Method   | Endpoint                          | Description           |
| -------- | --------------------------------- | --------------------- |
| `GET`    | `/api/admin/projects`             | List all projects     |
| `POST`   | `/api/admin/projects`             | Create a new project  |
| `DELETE` | `/api/admin/projects/:slug`       | Delete a project      |
| `GET`    | `/api/admin/updates`              | List all updates      |
| `POST`   | `/api/admin/updates`              | Register a new update |
| `PATCH`  | `/api/admin/updates/:id`          | Modify an update      |
| `DELETE` | `/api/admin/updates/:id`          | Delete an update      |
| `POST`   | `/api/admin/updates/:id/rollback` | Create a rollback     |
| `GET`    | `/api/admin/stats`                | Get analytics summary |
| `GET`    | `/api/admin/keys`                 | List API keys         |
| `POST`   | `/api/admin/keys`                 | Create an API key     |
| `DELETE` | `/api/admin/keys/:id`             | Delete an API key     |

## Client Configuration

Configure your Expo app to use OTAShip in `app.json`:

```json
{
  "expo": {
    "slug": "my-app",
    "runtimeVersion": "1",
    "updates": {
      "url": "https://your-server.com/api/my-app/manifest",
      "enabled": true
    }
  }
}
```

## Publishing Updates

Use the OTAShip CLI to publish updates.

### Quick Start

```bash
# Navigate to your Expo project
cd my-expo-app

# Initialize OTAShip config (one-time setup)
otaship init

# Publish an update
otaship
```

### Configuration

The CLI stores configuration in `otaship.json`:

```json
{
  "server": "https://your-server.com",
  "api": "YOUR_API_KEY",
  "channel": "production"
}
```

### CLI Options

| Flag            | Default      | Description                |
| --------------- | ------------ | -------------------------- |
| `--project`     | `.`          | Path to Expo project       |
| `--server`      | from config  | OTAShip server URL         |
| `--api`         | from config  | API key for authentication |
| `--channel`     | `production` | Update channel             |
| `--rollout`     | `100`        | Rollout percentage (0-100) |
| `--skip-export` | `false`      | Skip `expo export` step    |

### Examples

```bash
# Publish with config file (recommended)
otaship

# Override server and channel
otaship --server https://otaship.onrender.com --channel staging

# Publish with 50% rollout
otaship --rollout 50

# Skip export if already done
otaship --skip-export
```

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Expo Client   │────▶│  OTAShip API    │────▶│   Cloudinary    │
│   (Mobile App)  │     │   (Go + Gin)    │     │     (CDN)       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │  MongoDB Atlas  │
                        └─────────────────┘
```

## Project Structure

```
backend/
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # MongoDB repositories
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # Auth and CORS middleware
│   ├── models/            # Data models
│   ├── services/          # Business logic (signing, rollout)
│   └── storage/           # Cloudinary integration
└── updates/               # Local update storage (development)
```

## License

MIT
