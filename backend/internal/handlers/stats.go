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
			total  []database.GetTotalDownloadsByProjectRow
			recent []database.GetRecentDownloadsByProjectRow
		)

		g, ctx := errgroup.WithContext(r.Context())

		g.Go(func() error {
			var err error
			total, err = queries.GetTotalDownloadsByProject(ctx, projectId)
			return err
		})

		g.Go(func() error {
			var err error
			recent, err = queries.GetRecentDownloadsByProject(ctx, projectId)
			return err
		})

		if err := g.Wait(); err != nil {
			jsonError(w, "Failed to fetch project stats", http.StatusInternalServerError)
			return
		}

		var totalDownloads int64
		var recentDownloads int64
		platforms := make(map[string]int64)
		channels := make(map[string]int64)

		for _, r := range recent {
			recentDownloads += r.Count
			totalDownloads += r.Count
			platforms[r.Platform] += r.Count
			channels[r.Channel] += r.Count
		}

		totalDownloads += recentDownloads

		for _, r := range total {
			totalDownloads += r.Count
			platforms[r.Platform] += r.Count
			channels[r.Channel] += r.Count
		}

		type statItem struct {
			Platform string `json:"platform,omitempty"`
			Channel  string `json:"channel,omitempty"`
			Count    int64  `json:"count"`
		}

		byPlatform := make([]statItem, 0)
		for p, c := range platforms {
			byPlatform = append(byPlatform, statItem{Platform: p, Count: c})
		}

		byChannel := make([]statItem, 0)
		for ch, c := range channels {
			byChannel = append(byChannel, statItem{Channel: ch, Count: c})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_downloads":  totalDownloads,
			"recent_downloads": recentDownloads,
			"by_platform":      byPlatform,
			"by_channel":       byChannel,
		})
	}
}

func GetGlobalStats(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			total  []database.GetTotalDownloadStatsRow
			recent []database.GetGlobalRecentDownloadsRow
		)

		g, ctx := errgroup.WithContext(r.Context())

		g.Go(func() error {
			var err error
			total, err = queries.GetTotalDownloadStats(ctx)
			return err
		})

		g.Go(func() error {
			var err error
			recent, err = queries.GetGlobalRecentDownloads(ctx)
			return err
		})

		if err := g.Wait(); err != nil {
			jsonError(w, "Failed to fetch global stats", http.StatusInternalServerError)
			return
		}

		var totalDownloads int64
		var recentDownloads int64
		platforms := make(map[string]int64)
		channels := make(map[string]int64)

		for _, r := range recent {
			recentDownloads += r.Count
			totalDownloads += r.Count
			platforms[r.Platform] += r.Count
			channels[r.Channel] += r.Count
		}

		for _, r := range total {
			totalDownloads += r.Count
			platforms[r.Platform] += r.Count
			channels[r.Channel] += r.Count
		}

		type statItem struct {
			Platform string `json:"platform,omitempty"`
			Channel  string `json:"channel,omitempty"`
			Count    int64  `json:"count"`
		}

		byPlatform := make([]statItem, 0)
		for p, c := range platforms {
			byPlatform = append(byPlatform, statItem{Platform: p, Count: c})
		}

		byChannel := make([]statItem, 0)
		for ch, c := range channels {
			byChannel = append(byChannel, statItem{Channel: ch, Count: c})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_downloads":  totalDownloads,
			"recent_downloads": recentDownloads,
			"by_platform":      byPlatform,
			"by_channel":       byChannel,
		})
	}
}
