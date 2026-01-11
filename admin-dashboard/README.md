# OTAShip Admin Dashboard

Web dashboard for managing OTA updates, projects, and viewing analytics.

Built with Next.js and Tailwind CSS.

## Features

- **Project Management** — Create and manage multiple Expo projects
- **Update Control** — View, activate, deactivate, and delete updates
- **Rollout Management** — Adjust rollout percentages
- **Analytics** — View download statistics and adoption rates
- **API Keys** — Manage API keys for CLI access

## Setup

### Prerequisites

- Node.js 18+
- OTAShip backend running

### Installation

```bash
# Install dependencies
npm install

# Copy environment template
cp .env.example .env.local

# Configure your environment
# Edit .env.local and set NEXT_PUBLIC_API_URL
```

### Environment Variables

| Variable              | Description         | Example                        |
| --------------------- | ------------------- | ------------------------------ |
| `NEXT_PUBLIC_API_URL` | OTAShip backend URL | `https://otaship.onrender.com` |

### Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

### Production Build

```bash
npm run build
npm start
```

### Deploy to Vercel

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new)

Set the `NEXT_PUBLIC_API_URL` environment variable in your Vercel project settings.

## Pages

| Route       | Description                  |
| ----------- | ---------------------------- |
| `/`         | Dashboard home with overview |
| `/projects` | Project management           |
| `/releases` | Update/release management    |
| `/settings` | API keys and configuration   |
| `/login`    | Authentication               |

## Tech Stack

- Next.js 15
- React 19
- Tailwind CSS
- Recharts (for analytics charts)
- Lucide React (icons)

## License

MIT
