package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/utils"
	"golang.org/x/sync/errgroup"
)

func since() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time.Now().Add(-24 * time.Hour),
		Valid: true,
	}
}

func GetProjectStats(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "project_id")
		if id == "" {
			jsonError(w, "project_id parameter is required", http.StatusBadRequest)
			return
		}

		projectId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		var (
			total      int64
			recent     int64
			byPlatform []database.GetDownloadsByPlatformRow
			byChannel  []database.GetDownloadsByChannelRow
			byUpdate   []database.GetDownloadsByUpdateRow
		)

		g, ctx := errgroup.WithContext(r.Context())

		g.Go(func() error {
			var err error
			total, err = queries.GetTotalDownloadsByProject(ctx, projectId)
			return err
		})

		g.Go(func() error {
			var err error
			recent, err = queries.GetRecentDownloadsByProject(ctx, database.GetRecentDownloadsByProjectParams{
				ProjectID: projectId,
				Since:     since(),
			})
			return err
		})

		g.Go(func() error {
			var err error
			byPlatform, err = queries.GetDownloadsByPlatform(ctx, projectId)
			return err
		})

		g.Go(func() error {
			var err error
			byChannel, err = queries.GetDownloadsByChannel(ctx, projectId)
			return err
		})

		g.Go(func() error {
			var err error
			byUpdate, err = queries.GetDownloadsByUpdate(ctx, projectId)
			return err
		})

		if err := g.Wait(); err != nil {
			jsonError(w, "Failed to fetch project stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_downloads":  total,
			"recent_downloads": recent,
			"by_platform":      byPlatform,
			"by_channel":       byChannel,
			"by_update":        byUpdate,
		})
	}
}

func GetGlobalStats(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			total      int64
			recent     int64
			byPlatform []database.GetGlobalDownloadsByPlatformRow
			byChannel  []database.GetGlobalDownloadsByChannelRow
		)

		g, ctx := errgroup.WithContext(r.Context())

		g.Go(func() error {
			var err error
			total, err = queries.GetGlobalTotalDownloads(ctx)
			return err
		})

		g.Go(func() error {
			var err error
			recent, err = queries.GetGlobalRecentDownloads(ctx, since())
			return err
		})

		g.Go(func() error {
			var err error
			byPlatform, err = queries.GetGlobalDownloadsByPlatform(ctx)
			return err
		})

		g.Go(func() error {
			var err error
			byChannel, err = queries.GetGlobalDownloadsByChannel(ctx)
			return err
		})

		if err := g.Wait(); err != nil {
			jsonError(w, "Failed to fetch global stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_downloads":  total,
			"recent_downloads": recent,
			"by_platform":      byPlatform,
			"by_channel":       byChannel,
		})
	}
}
