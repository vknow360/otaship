### OTAShip Backend: Comprehensive Architectural and Functional Analysis

#### 🏗️ Architecture Overview
The OTAShip backend is a high-performance **Go (1.25)** service designed to manage and serve **Expo OTA (Over-The-Air) updates**. It follows a standard clean architecture pattern with a clear separation between routing, business logic (handlers), and data persistence.

-   **Web Framework**: `chi/v5` for lightweight, idiomatic routing and middleware support.
-   **Database**: **PostgreSQL** with `pgxpool` for connection pooling.
-   **Data Mapping**: `sqlc` is used to generate type-safe Go code from raw SQL queries, ensuring performance and reliability.
-   **Asset Storage**: **Cloudinary** integration for storing and serving update bundles (JS bundles and assets).
-   **Authentication**: Custom API key-based authentication for CLI/Admin tools, using **bcrypt** for hashing.

---

#### 🛠️ Functional Features

##### 1. Expo Manifest API (`/api/manifest/{projectId}`)
This is the core of the OTA update system. It implements the **Expo Manifest Protocol (v1)**:
-   **Protocol Support**: Handles both legacy (v0) and modern (v1) Expo protocols using `multipart/mixed` responses.
-   **Smart Filtering**: Delivers the latest active update based on `runtime-version`, `platform` (iOS/Android), and `channel` (production/staging/beta).
-   **Rollout Control**: Supports partial rollouts via a `rollout_percentage` check before serving updates.
-   **Optimized Delivery**: Uses ETag-like checks via `expo-current-update-id` to prevent redundant downloads.

##### 2. Project & Update Management
-   **Project Scoping**: Multi-tenant design where updates and assets are scoped to specific projects.
-   **Lifecycle Management**:
    -   `CreateProject`: Generates unique slugs and secure API keys.
    -   `CreateUpdate`: Automatically deactivates older updates in the same channel/platform when a new one is published.
    -   `Rollback Support`: Database schema includes `is_rollback` flag for quick version reversion.
    -   `Rollout Adjustments`: Allows dynamic patching of rollout percentages for staged releases.

##### 3. Asset & Bundle Processing
-   **ZIP Handling**: Backend accepts `.zip` bundles from the CLI, extracts them, and parses `metadata.json` to identify platform-specific assets.
-   **Integrity Verification**: Calculates **SHA256** hashes for every file to ensure bundle integrity during client downloads.
-   **MIME Inference**: Intelligently detects content types for JS bundles (`.hbc`, `.js`) and assets.

---

#### 🚧 Remaining & Expected Features

-   **Analytics Ingestion**: The database supports `download_events` (tracking unique devices, platforms, and channels), but there is currently no public API endpoint to receive these events from the Expo client.
-   **Advanced Key Management**: While a dedicated `api_keys` table exists, the logic for managing multiple keys per project in `apikeys.go` is currently commented out and non-functional.
-   **Security Hardening**: The `AdminOnly` middleware is a placeholder that allows all requests. Implementing a real authentication layer (like JWT or Session-based) is a pending requirement.
-   **Automated Asset Cleanup**: When a project or update is deleted from the database, the corresponding files remain in Cloudinary storage. A background worker or hook for storage cleanup is missing.

---

#### 🐛 Potential Bugs & Improvements

-   **Silent UUID Failures**: In `CheckForUpdates`, UUID parsing errors are ignored (`_`), which could cause the backend to search for updates using a zero-UUID (`00000000-0000-0000-0000-000000000000`) instead of returning a `400 Bad Request`.
-   **Context Handling**: Background goroutines (e.g., updating `last_used_at` for API keys) use `context.Background()` instead of a derived context, making them harder to trace or cancel during server shutdown.
-   **Asset Type Detection**: The backend relies on `http.DetectContentType` and manual extension checks. For Hermes bytecode (`.hbc`), it might benefit from more robust validation to ensure clients receive the correct binary format.

---

#### 🚀 Key Technologies
| Component | Technology |
| :--- | :--- |
| **Language** | Go 1.25 |
| **Router** | go-chi/chi |
| **ORM/SQL** | sqlc |
| **Driver** | jackc/pgx/v5 |
| **Storage** | Cloudinary |
| **Crypto** | golang.org/x/crypto/bcrypt |