package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/storage"
	"github.com/vknow360/otaship/backend/internal/utils"
)

type CreateUpdateParams struct {
	ProjectID         string `json:"project_id"`
	RuntimeVersion    string `json:"runtime_version"`
	Channel           string `json:"channel"`
	RolloutPercentage int32  `json:"rollout_percentage"`
	Platform          string `json:"platform"`
	IsRollback        bool   `json:"is_rollback"`
	Message           string `json:"message"`
}

type UpdateResponse struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	RuntimeVersion    string `json:"runtime_version"`
	Channel           string `json:"channel"`
	RolloutPercentage int32  `json:"rollout_percentage"`
	Platform          string `json:"platform"`
	IsActive          bool   `json:"is_active"`
	IsRollback        bool   `json:"is_rollback"`
	Message           string `json:"message"`
	CreatedAt         int64  `json:"created_at"`
}

func toUpdateResponse(u database.Update) UpdateResponse {
	return UpdateResponse{
		ID:                u.ID.String(),
		ProjectID:         u.ProjectID.String(),
		RuntimeVersion:    u.RuntimeVersion,
		Channel:           u.Channel,
		RolloutPercentage: u.RolloutPercentage,
		Platform:          u.Platform,
		IsActive:          u.IsActive,
		IsRollback:        u.IsRollback,
		Message:           u.Message.String,
		CreatedAt:         u.CreatedAt.Time.UnixMilli(),
	}
}

// Admin-scoped: list updates with optional project_id filter and pagination
func ListUpdates(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id := query.Get("project_id")

		limitStr := query.Get("limit")
		offsetStr := query.Get("offset")

		limit := int32(math.MaxInt32)
		offset := int32(0)

		if limitStr != "" {
			if l, err := utils.ParseInt32(limitStr); err == nil {
				limit = l
			}
		}
		if offsetStr != "" {
			if o, err := utils.ParseInt32(offsetStr); err == nil {
				offset = o
			}
		}

		var updates []database.Update
		var total int64
		var err error

		if id != "" {
			projectId, err := utils.ParseUUID(id)
			if err != nil {
				jsonError(w, "Invalid project ID", http.StatusBadRequest)
				return
			}
			updates, err = queries.ListUpdatesByProject(r.Context(), database.ListUpdatesByProjectParams{
				ProjectID: projectId,
				Limit:     limit,
				Offset:    offset,
			})
			if err == nil {
				total, _ = queries.GetUpdatesCountByProject(r.Context(), projectId)
			}
		} else {
			updates, err = queries.ListUpdatesPaginated(r.Context(), database.ListUpdatesPaginatedParams{
				Limit:  limit,
				Offset: offset,
			})
			if err == nil {
				total, _ = queries.GetUpdatesCount(r.Context())
			}
		}

		if err != nil {
			jsonError(w, "Failed to fetch updates", http.StatusInternalServerError)
			return
		}

		if updates == nil {
			updates = []database.Update{}
		}

		resp := make([]UpdateResponse, len(updates))
		for i, u := range updates {
			resp[i] = toUpdateResponse(u)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"updates": resp,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	}
}

// Project-scoped: list updates for the authenticated project
func ListProjectUpdates(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := utils.GetProjectId(r.Context())

		updates, err := queries.ListUpdatesByProject(r.Context(), database.ListUpdatesByProjectParams{
			ProjectID: projectId,
			Limit:     math.MaxInt32,
			Offset:    0,
		})
		if err != nil {
			jsonError(w, "Failed to fetch updates", http.StatusInternalServerError)
			return
		}
		if updates == nil {
			updates = []database.Update{}
		}

		resp := make([]UpdateResponse, len(updates))
		for i, u := range updates {
			resp[i] = toUpdateResponse(u)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// Admin-scoped: delete any update
func DeleteUpdate(queries *database.Queries, providers map[string]storage.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "update_id")
		if id == "" {
			jsonError(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		updateId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}

		// Fetch the update first so we can invalidate the right project cache
		updateRow, err := queries.GetUpdateByID(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Update not found", http.StatusNotFound)
			return
		}

		updateAssets, _ := queries.GetAssetsByUpdateID(r.Context(), updateId)

		for _, asset := range updateAssets {
			count, err := queries.CountOtherAssetReferences(r.Context(), database.CountOtherAssetReferencesParams{
				Key:      asset.Key,
				UpdateID: updateId,
			})

			if err == nil && count == 0 {
				targetProvider, exists := providers[asset.StorageProvider]
				if !exists {
					slog.WarnContext(r.Context(), "Storage provider not found", slog.String("key", asset.Key))
					continue
				}
				go func(provider storage.Provider, k, mime string) {
					_ = provider.Delete(context.Background(), k, mime)
				}(targetProvider, asset.Key, asset.MimeType)
			}
		}

		err = queries.DeleteUpdate(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Failed to delete update", http.StatusInternalServerError)
			return
		}

		InvalidateManifestCache(updateRow.ProjectID.String())
		w.WriteHeader(http.StatusNoContent)
	}
}

// Project-scoped: delete update owned by the authenticated project
func DeleteProjectUpdate(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "update_id")
		if id == "" {
			jsonError(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		updateId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}

		update, err := queries.GetUpdateByID(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Update not found", http.StatusNotFound)
			return
		}

		if update.ProjectID != utils.GetProjectId(r.Context()) {
			jsonError(w, "Update does not belong to this project", http.StatusForbidden)
			return
		}

		err = queries.DeleteUpdate(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Failed to delete update", http.StatusInternalServerError)
			return
		}

		InvalidateManifestCache(update.ProjectID.String())
		w.WriteHeader(http.StatusNoContent)
	}
}

func CreateUpdate(pool *pgxpool.Pool, queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update CreateUpdateParams
		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if update.ProjectID == "" || update.Channel == "" || update.Platform == "" || update.RuntimeVersion == "" {
			jsonError(w, "Missing required fields", http.StatusBadRequest)
			return
		}
		projectId, err := utils.ParseUUID(update.ProjectID)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		if projectId != utils.GetProjectId(r.Context()) {
			jsonError(w, "Project ID does not match the authenticated user", http.StatusForbidden)
			return
		}

		if matched := regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,32}$`).Match([]byte(update.Channel)); !matched {
			jsonError(w, "Invalid channel name. Must start with a letter or number and contain only letters, numbers, hyphens, and underscores, with a maximum length of 32 characters.", http.StatusBadRequest)
			return
		}
		if matched := regexp.MustCompile(`^\d+(\.\d+)*(-[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*)?$`).Match([]byte(update.RuntimeVersion)); !matched {
			jsonError(w, "Invalid runtime version. Must be a valid semver or simple number (e.g., 2, 1.0, 1.0.0-alpha.1).", http.StatusBadRequest)
			return
		}
		if matched := regexp.MustCompile(`^(ios|android|all)$`).Match([]byte(update.Platform)); !matched {
			jsonError(w, "Invalid platform. Must be either ios or android or all.", http.StatusBadRequest)
			return
		}
		if update.RolloutPercentage < 0 || update.RolloutPercentage > 100 {
			jsonError(w, "Invalid rollout percentage. Must be between 0 and 100.", http.StatusBadRequest)
			return
		}

		tx, err := pool.Begin(r.Context())
		if err != nil {
			jsonError(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}

		defer tx.Rollback(r.Context())

		qtx := queries.WithTx(tx)

		var messageText pgtype.Text
		if update.Message != "" {
			messageText = pgtype.Text{String: update.Message, Valid: true}
		}

		createUpdate, err := qtx.CreateUpdate(r.Context(), database.CreateUpdateParams{
			ProjectID:         projectId,
			RuntimeVersion:    update.RuntimeVersion,
			Channel:           update.Channel,
			RolloutPercentage: update.RolloutPercentage,
			Platform:          update.Platform,
			IsActive:          false,
			IsRollback:        false,
			Message:           messageText,
		})
		if err != nil {
			jsonError(w, "Failed to create update", http.StatusInternalServerError)
			return
		}

		err = tx.Commit(r.Context())
		if err != nil {
			jsonError(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toUpdateResponse(createUpdate))
	}
}

type UpdateRolloutRequest struct {
	RolloutPercentage int32 `json:"rollout_percentage"`
}

func UpdateRolloutPercentage(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "update_id")
		if id == "" {
			jsonError(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		updateId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}

		var updateRollout UpdateRolloutRequest
		err = json.NewDecoder(r.Body).Decode(&updateRollout)
		if err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if updateRollout.RolloutPercentage < 0 || updateRollout.RolloutPercentage > 100 {
			jsonError(w, "Rollout percentage must be between 0 and 100", http.StatusBadRequest)
			return
		}

		err = queries.UpdateRolloutPercentage(r.Context(), database.UpdateRolloutPercentageParams{
			ID:                updateId,
			RolloutPercentage: updateRollout.RolloutPercentage,
		})
		if err != nil {
			jsonError(w, "Failed to update rollout percentage", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func CreateRollback(pool *pgxpool.Pool, queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateIdStr := chi.URLParam(r, "update_id")
		if updateIdStr == "" {
			jsonError(w, "Update ID is required", http.StatusBadRequest)
			return
		}

		updateId, err := utils.ParseUUID(updateIdStr)
		if err != nil {
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}

		original, err := queries.GetUpdateByID(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Update not found", http.StatusNotFound)
			return
		}

		ctxProjectId := utils.GetProjectId(r.Context())
		if ctxProjectId.Valid && original.ProjectID != ctxProjectId {
			jsonError(w, "Update does not belong to this project", http.StatusForbidden)
			return
		}

		tx, err := pool.Begin(r.Context())
		if err != nil {
			jsonError(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(r.Context())

		qtx := queries.WithTx(tx)

		err = qtx.DeactivateUpdates(r.Context(), database.DeactivateUpdatesParams{
			ProjectID:      original.ProjectID,
			Channel:        original.Channel,
			RuntimeVersion: original.RuntimeVersion,
			Platform:       original.Platform,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to deactivate updates during rollback",
				slog.String("update_id", updateIdStr),
				slog.Any("error", err),
			)
			jsonError(w, "Failed to deactivate updates", http.StatusInternalServerError)
			return
		}

		rollback, err := qtx.CreateUpdate(r.Context(), database.CreateUpdateParams{
			ProjectID:         original.ProjectID,
			RuntimeVersion:    original.RuntimeVersion,
			Channel:           original.Channel,
			RolloutPercentage: 100,
			Platform:          original.Platform,
			IsActive:          true,
			IsRollback:        true,
			Message:           pgtype.Text{String: "Rollback from " + updateIdStr, Valid: true},
		})
		if err != nil {
			jsonError(w, "Failed to create rollback", http.StatusInternalServerError)
			return
		}

		err = qtx.CloneAssets(r.Context(), database.CloneAssetsParams{
			UpdateID:   updateId,
			UpdateID_2: rollback.ID,
		})
		if err != nil {
			jsonError(w, "Failed to clone assets", http.StatusInternalServerError)
			return
		}

		err = tx.Commit(r.Context())
		if err != nil {
			jsonError(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		go InvalidateManifestCache(original.ProjectID.String())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(toUpdateResponse(rollback))
	}
}
