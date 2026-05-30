# OTAShip Admin Dashboard

A beautiful, responsive web interface for managing your self-hosted OTAShip instance. Built with SvelteKit, this dashboard provides complete visibility and control over your OTA update pipeline.

## Features

- **Project Management:** Create new projects and manage existing ones.
- **Update History:** View a timeline of all published updates per project, including channels, platforms, and release dates.
- **Access Control:** Generate, view, and revoke API keys used by the OTAShip CLI.
- **Rollbacks:** Instantly revert broken releases directly from the web interface.
- **Monitoring:** View configuration details, update metadata, and storage backend settings.

## Prerequisites

- Node.js 18 or higher
- `npm` or `pnpm` (pnpm is recommended)
- A running instance of the OTAShip backend

## Setup

### 1. Environment Variables

Create a `.env` file in the root of the `admin-dashboard` directory by copying the example file:

```bash
cp .env.example .env
```

Configure the connection to your OTAShip backend:

| Variable | Description | Default |
|----------|-------------|---------|
| `PUBLIC_API_URL` | The base URL of your OTAShip backend API | `http://localhost:8080` |

### 2. Installation

Install the dependencies:

```bash
npm install
# or
pnpm install
```

### 3. Development Server

Start the local development server:

```bash
npm run dev
# or
pnpm dev
```
The dashboard will be available at `http://localhost:5173`. 

To log in, use the plaintext password that corresponds to the `ADMIN_TOKEN_HASH` you configured in your backend's environment variables.

## Building for Production

To create an optimized production build of the dashboard:

```bash
npm run build
```

By default, the dashboard uses SvelteKit's `adapter-auto`. If you are deploying to a specific environment (like Node.js, Vercel, or Cloudflare Pages), you may need to install the corresponding SvelteKit adapter and update your `svelte.config.js`. 

For example, to run the dashboard as a standalone Node.js server, switch to `@sveltejs/adapter-node`.
