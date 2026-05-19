package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/handlers"
	"github.com/vknow360/otaship/backend/internal/logger"
	mid "github.com/vknow360/otaship/backend/internal/middleware"
	"github.com/vknow360/otaship/backend/internal/storage"
)

var accessToken string

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Warning: .env file not found. Using environment variables.")
	}

	logger.InitLogger(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_FORMAT"))

	accessToken = os.Getenv("ADMIN_TOKEN_HASH")
	if accessToken == "" {
		panic("ADMIN_TOKEN_HASH is not set")
	}

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	if err = runMigrations(dbURL); err != nil {
		slog.Error("Migration failed:", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	queries := database.New(db)

	providers := make(map[string]storage.Provider)

	s3, err := storage.NewS3Provider()
	if err != nil {
		slog.Error("Failed to connect to S3", slog.String("error", err.Error()))
	} else {
		slog.Info("Connected to S3")
		providers["s3"] = s3
	}

	cld, err := storage.NewCloudinaryProvider()
	if err != nil {
		slog.Error("Failed to connect to Cloudinary", slog.String("error", err.Error()))
	} else {
		slog.Info("Connected to Cloudinary")
		providers["cloudinary"] = cld
	}

	if len(providers) == 0 {
		panic("No storage provider configured")
	}

	setDefaultProvider(queries, providers)

	r := chi.NewRouter()
	r.Use(logger.Middleware)
	r.Use(middleware.Recoverer)

	// CORS
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 0 || allowedOrigins[0] == "" {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Authorization", "X-API-Key",
			"expo-platform", "expo-runtime-version", "expo-channel-name",
			"expo-protocol-version", "expo-expect-signature",
			"expo-current-update-id", "expo-embedded-update-id",
		},
		ExposedHeaders: []string{
			"expo-protocol-version", "expo-sfv-version", "expo-signature",
		},
		MaxAge: 300,
	}))

	r.Get("/health", handlers.HealthCheck(db))

	r.Get("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	r.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<title>OTAship API Swagger Docs</title>
		<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
		</head>
		<body>
		<div id="swagger-ui"></div>
		<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
		<script>
			window.onload = () => {
				window.ui = SwaggerUIBundle({
					url: '/api/openapi.yaml',
					dom_id: '#swagger-ui',
				});
			};
		</script>
		</body>
		</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	r.Mount("/api", apiRouter(queries))
	r.Mount("/api/project", projectRouter(db, queries, providers))
	r.Mount("/api/admin", adminRouter(db, queries, providers))

	startAggregationJob(db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("Server starting", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func apiRouter(queries *database.Queries) http.Handler {
	r := chi.NewRouter()

	limiter := httprate.NewRateLimiter(10, time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP), httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
	}))
	r.Use(limiter.Handler)
	r.Get("/manifest/{project_id}", handlers.CheckForUpdates(queries))
	r.Get("/validate-key", handlers.ValidateAPIKey(queries))

	return r
}

func projectRouter(db *pgxpool.Pool, queries *database.Queries, providers map[string]storage.Provider) http.Handler {
	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(30, time.Minute))
	r.Use(mid.ProjectKeyOnly(queries))
	r.Get("/me", handlers.GetMe(queries))

	r.Post("/{project_id}/updates/{update_id}/upload", handlers.UploadAsset(db, queries, providers))

	r.Post("/updates", handlers.CreateUpdate(db, queries))
	r.Get("/updates", handlers.ListProjectUpdates(queries))
	r.Delete("/updates/{update_id}", handlers.DeleteProjectUpdate(queries, providers))
	r.Post("/updates/{update_id}/rollback", handlers.CreateRollback(db, queries))
	return r
}

func adminRouter(db *pgxpool.Pool, queries *database.Queries, providers map[string]storage.Provider) http.Handler {
	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(100, time.Minute))
	r.Use(mid.AdminOnly(accessToken))

	r.Get("/verify", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	r.Get("/projects", handlers.GetProjects(queries))
	r.Post("/projects", handlers.CreateProject(queries))
	r.Get("/projects/{project_id}", handlers.GetProjectByID(queries))
	r.Patch("/projects/{project_id}", handlers.UpdateProject(queries))
	r.Delete("/projects/{project_id}", handlers.DeleteProject(queries))
	r.Get("/projects/{project_id}/stats", handlers.GetProjectStats(queries))
	r.Post("/projects/{project_id}/keys", handlers.CreateAPIKey(queries))
	r.Get("/projects/{project_id}/keys", handlers.ListAPIKeys(queries))
	r.Delete("/projects/{project_id}/keys/{key_id}", handlers.DeleteAPIKey(queries))

	r.Get("/updates", handlers.ListUpdates(queries))
	r.Get("/updates/{update_id}", handlers.GetUpdate(queries))
	r.Get("/updates/{update_id}/assets", handlers.ListUpdateAssets(queries))
	r.Patch("/updates/{update_id}/rollout", handlers.UpdateRolloutPercentage(queries))
	r.Delete("/updates/{update_id}", handlers.DeleteUpdate(queries, providers))
	r.Post("/updates/{update_id}/rollback", handlers.CreateRollback(db, queries))

	r.Get("/settings", handlers.GetSettings(queries, providers))
	r.Put("/settings", handlers.UpdateSetting(queries))
	r.Get("/settings/storage/usage", handlers.GetStorageUsage(providers))
	r.Get("/settings/{key}", handlers.GetSetting(queries))

	r.Get("/stats", handlers.GetGlobalStats(queries))

	return r
}

func runMigrations(dbURL string) error {
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func startAggregationJob(pool *pgxpool.Pool) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			// Run immediately on startup, then on every tick
			prune(pool)
			<-ticker.C
		}
	}()
}

func prune(pool *pgxpool.Pool) {
	ctx := context.Background()
	cutoff := time.Now().UTC().Truncate(24 * time.Hour)

	var pgCutoff pgtype.Timestamptz
	pgCutoff.Time = cutoff
	pgCutoff.Valid = true

	tx, err := pool.Begin(ctx)
	if err != nil {
		slog.Error("Failed to begin transaction", slog.Any("error", err))
		return
	}
	defer tx.Rollback(ctx)

	qtx := database.New(tx)

	err = qtx.AggregateDownloadEvents(ctx, pgCutoff)
	if err != nil {
		slog.Error("Aggregation failed", slog.Any("error", err))
		return
	}

	err = qtx.DeleteAggregatedEvents(ctx, pgCutoff)
	if err != nil {
		slog.Error("Failed to delete aggregated events", slog.Any("error", err))
	} else {
		slog.Info("Aggregation complete", slog.Time("cutoff", cutoff))
	}

	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("Failed to commit transaction", slog.Any("error", err))
	}
}

func setDefaultProvider(queries *database.Queries, providers map[string]storage.Provider) {
	ctx := context.Background()

	_, err := queries.GetSetting(ctx, "storage_provider")
	if err != nil {
		defaultProvider := ""
		if providers["cloudinary"] != nil {
			defaultProvider = "cloudinary"
		} else if providers["s3"] != nil {
			defaultProvider = "s3"
		}
		queries.UpdateSetting(ctx, database.UpdateSettingParams{
			Key:   "storage_provider",
			Value: defaultProvider,
		})
	}
}
