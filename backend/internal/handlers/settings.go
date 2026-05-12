package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/storage"
)

func GetSettings(queries *database.Queries, providers map[string]storage.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := queries.GetSettings(r.Context())
		if err != nil {
			slog.Error("Failed to get settings", "error", err)
			jsonError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var resp = make(map[string]any, len(settings))
		for _, setting := range settings {
			resp[setting.Key] = setting.Value
		}

		var providersList []string
		for provider := range providers {
			providersList = append(providersList, provider)
		}
		resp["providers"] = providersList
		json.NewEncoder(w).Encode(resp)
	}
}

func GetSetting(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		if key == "" {
			jsonError(w, "key is required", http.StatusBadRequest)
			return
		}
		setting, err := queries.GetSetting(r.Context(), key)
		if err != nil {
			slog.Error("Failed to get setting", "error", err)
			jsonError(w, "internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(setting)
	}
}

func UpdateSetting(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var setting database.UpdateSettingParams
		if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
			slog.Error("Failed to decode setting", "error", err)
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := queries.UpdateSetting(r.Context(), setting); err != nil {
			slog.Error("Failed to update setting", "error", err)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetStorageUsage(providers map[string]storage.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := make(map[string]any)
		for provider, storage := range providers {
			usage, err := storage.Usage(r.Context())
			if err != nil {
				slog.Error("Failed to get storage usage", "provider", provider, "error", err)
				continue
			}
			stats[provider] = usage
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
