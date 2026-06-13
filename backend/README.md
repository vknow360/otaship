# OTAShip — Backend

The Go API server that powers OTAShip. Handles update uploads, serves Expo-compatible manifests, manages rollouts, and talks to your storage provider.

→ [Back to main README](../README.md)

## Tech Stack

| | |
|---|---|
| **Language** | Go 1.25+ |
| **Router** | [go-chi/chi](https://github.com/go-chi/chi) |
| **Database** | PostgreSQL 16+ via [pgx](https://github.com/jackc/pgx) connection pool |
| **Query Generation** | [sqlc](https://sqlc.dev/) — type-safe SQL, no ORM |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) — runs automatically on startup |
| **Storage** | Pluggable: AWS S3 / MinIO / Cloudinary |

## What It Does

The backend is the central piece of OTAShip:

- **For the CLI:** Receives update bundles via `X-API-Key` authenticated uploads
- **For the Dashboard:** Provides admin REST APIs (projects, updates, API keys, settings) secured with bearer tokens
- **For the Expo App:** Serves signed manifests following the Expo Updates protocol, with percentage-based rollouts and channel targeting
- **Internally:** Auto-runs database migrations, aggregates download stats daily, and serves interactive Swagger docs

## Local Development

> Requires Go 1.25+ and a running PostgreSQL instance.

```bash
cd backend
cp .env.example .env
# Edit .env with your database URL and storage credentials
go run ./cmd/server
```

Migrations run automatically on startup. The server starts on `http://localhost:8080`.

Alternatively, you can run the backend and database via Docker Compose from the repository root:

```bash
# From the root directory
docker compose -f docker-compose.dev.yml up -d
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | ✅ | PostgreSQL connection string |
| `ADMIN_TOKEN_HASH` | ✅ | SHA-256 hash of your admin password |
| `PORT` | | Server port (default: `8080`) |
| `S3_ACCESS_KEY` | ¹ | AWS/MinIO access key |
| `S3_SECRET_ACCESS_KEY` | ¹ | AWS/MinIO secret key |
| `S3_REGION` | ¹ | S3 region |
| `S3_BUCKET_NAME` | ¹ | S3 bucket name |
| `S3_ENDPOINT` | ¹ | Custom endpoint (for MinIO) |
| `S3_BASE_PATH` | | Prefix path inside the bucket |
| `CLOUDINARY_CLOUD_NAME` | ² | Cloudinary cloud name |
| `CLOUDINARY_API_KEY` | ² | Cloudinary API key |
| `CLOUDINARY_API_SECRET` | ² | Cloudinary API secret |
| `EXPO_PRIVATE_KEY` | | RSA private key for manifest code signing |
| `ALLOWED_ORIGINS` | | CORS origins, comma-separated (default: `*`) |
| `LOG_FORMAT` | | `text` or `json` (default: `text`) |
| `LOG_LEVEL` | | `debug`, `info`, `warn`, `error` (default: `debug`) |

> ¹ Required if using S3/MinIO as storage provider
> ² Required if using Cloudinary as storage provider
> At least one storage provider must be configured.

## API Documentation

Interactive Swagger docs are available at:

```
http://localhost:8080/api/docs
```

The raw OpenAPI spec is served at `/api/openapi.yaml`.

## Project Structure

```
backend/
├── cmd/server/          # Entry point, router setup, startup banner
├── internal/
│   ├── database/        # sqlc-generated Go code (do not edit manually)
│   ├── handlers/        # HTTP route handlers (admin, project, manifest)
│   ├── logger/          # Structured logging (slog) setup + middleware
│   ├── middleware/       # Auth (admin bearer, API key), CORS, rate limiting
│   ├── storage/         # Storage provider interfaces (S3, Cloudinary)
│   └── utils/           # Shared helpers
├── migrations/          # PostgreSQL schema migration files
├── queries/             # Raw SQL queries (input for sqlc)
├── openapi.yaml         # API specification
└── sqlc.yaml            # sqlc configuration
```

## Database Workflow

OTAShip uses `sqlc` instead of an ORM — you write SQL, and sqlc generates type-safe Go code.

To make schema changes:

1. Add a new migration file in `migrations/`
2. Update queries in `queries/`
3. Run `sqlc generate` to regenerate `internal/database/`

> Never edit files in `internal/database/` directly — they're overwritten on every `sqlc generate`.