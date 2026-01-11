# OTAShip

Self-hosted OTA (Over-The-Air) update server for Expo and React Native applications.

Skip the app store review process and deliver updates to your users instantly.

## Overview

OTAShip is a complete solution for managing OTA updates, consisting of:

| Component                            | Description                                                            |
| ------------------------------------ | ---------------------------------------------------------------------- |
| [backend](./backend)                 | Go server handling update distribution, analytics, and CDN integration |
| [admin-dashboard](./admin-dashboard) | Next.js web dashboard for managing updates and projects                |
| [otaship-cli](./otaship-cli)         | Command-line tool for publishing updates                               |
| [expo-client](./expo-client)         | Example Expo app demonstrating OTA integration                         |

## Features

- **Instant Updates** — Push JavaScript bundle updates without app store review
- **Multi-Project** — Manage multiple Expo apps from a single server
- **Gradual Rollouts** — Release updates to a percentage of users
- **Multiple Channels** — Separate production, staging, and beta releases
- **Analytics** — Track download counts and update adoption
- **CDN Support** — Cloudinary integration for global asset delivery
- **Code Signing** — Optional RSA-SHA256 manifest signing

## Quick Start

### 1. Deploy the Backend

```bash
cd backend
cp .env.example .env
# Configure MongoDB and Cloudinary credentials
go run .
```

Or deploy to Render with one click:

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

### 2. Set Up the Dashboard

```bash
cd admin-dashboard
npm install
cp .env.example .env.local
# Set NEXT_PUBLIC_API_URL to your backend URL
npm run dev
```

### 3. Configure Your Expo App

Update `app.json`:

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

### 4. Publish Updates

```bash
cd my-expo-app
otaship init
otaship
```

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Expo Client   │────▶│  OTAShip API    │────▶│   Cloudinary    │
│   (Mobile App)  │     │   (Go + Gin)    │     │     (CDN)       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       │                       │                       │
       ▼                       ▼                       ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    MongoDB      │     │  Admin Dashboard│     │   OTAShip CLI   │
│   (Metadata)    │     │   (Next.js)     │     │   (Go Binary)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

## Documentation

- [Backend Setup](./backend/README.md) — Server configuration and API reference
- [CLI Usage](./otaship-cli/README.md) — Publishing updates from command line
- [Dashboard Guide](./admin-dashboard/README.md) — Managing projects and updates

## Requirements

- Go 1.21+
- Node.js 18+
- MongoDB Atlas (or local MongoDB)
- Cloudinary account (optional, for CDN)

## License

MIT
