# OTAShip Backend

The high-performance Go service that implements the Expo Updates protocol, manages OTA updates, and powers the Admin Dashboard.

## Prerequisites

To run or build the backend, you need:
- Go 1.25+
- PostgreSQL 16+
- Storage provider account (Cloudinary or AWS S3/MinIO)

## Configuration

The backend relies on environment variables. Copy `.env.example` to `.env` and configure the following variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | **Yes** | PostgreSQL connection string (e.g., `postgresql://user:pass@host:5432/dbname?sslmode=disable`) |
| `ADMIN_TOKEN_HASH` | **Yes** | A bcrypt hash of the secret token used to log into the Admin Dashboard |
| `CLOUDINARY_*` | One of | `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`, `CLOUDINARY_CLOUD_NAME` (if using Cloudinary) |
| `S3_*` | One of | `S3_ACCESS_KEY`, `S3_SECRET_ACCESS_KEY`, `S3_REGION`, `S3_BUCKET_NAME` (if using S3) |
| `PORT` | No | Port to run the server on (default: `8080`) |
| `ALLOWED_ORIGINS` | No | CORS allowed origins (default: `*`) |

> **Pro-Tip for `ADMIN_TOKEN_HASH`:** 
> You can generate a bcrypt hash for your chosen password using any online bcrypt generator or CLI tool.

## Running Locally

1. Start your local PostgreSQL instance and create the `otaship` database.
2. Ensure your `.env` file is fully configured.
3. Start the server:
   ```bash
   go run ./cmd/server
   ```
   *(The server will automatically execute database migrations on startup via the `golang-migrate` library.)*

## Running with Docker

From the root of the repository, you can run the entire backend and database stack using Docker Compose:

```bash
docker-compose up -d
```

## API Overview

The backend exposes several sets of REST endpoints utilizing the `go-chi` router:

- **Expo Client API (`/api/manifest/{project_id}`)**: Handles device requests for updates, serving standard or multipart Expo manifests based on the client protocol version.
- **Project API (`/api/project/*`)**: Endpoints used by the OTAShip CLI to publish updates, upload asset bundles, and manage project metadata. Secured via per-project API keys.
- **Admin API (`/api/admin/*`)**: Endpoints used by the Admin Dashboard to manage all projects, view analytics, and control storage configurations. Secured via the `ADMIN_TOKEN_HASH`.
- **Validation**: `/api/validate-key` and `/api/project/me` for CLI authentication.

## Key Design Decisions

- **SQLC**: Generates type-safe Go code directly from SQL queries for performance and compile-time safety.
- **Smart Protocol Negotiation**: Can serve both standard JSON (Protocol 0) and `multipart/mixed` (Protocol 1) manifests depending on what the Expo client supports.
- **Asset Hashing**: Computes SHA-256 hashes of all JS bundles and assets uploaded via the CLI to ensure integrity on the client side.
- **ETag Caching**: Heavily caches manifests and respects `expo-current-update-id` headers to return `304 Not Modified`, saving massive amounts of bandwidth.

## Known Limitations

- **Analytics Ingestion**: While the database supports `download_events` for tracking unique devices, the public endpoint for clients to report successful downloads is currently not implemented.
- **Multi-Key Management**: The database schema supports multiple API keys per project, but the current API implementation only actively manages one primary key per project.
- **Storage Cleanup**: Deleting an update from the database does not currently trigger a background job to delete the corresponding assets from S3 or Cloudinary.