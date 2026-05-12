package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/utils"
)

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	APIKey    string `json:"api_key"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	KeySuffix string `json:"key_suffix"`
	CreatedAt int64  `json:"created_at"`
	LastUsed  int64  `json:"last_used"`
}

type ListAPIKeysResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	KeySuffix string `json:"key_suffix"`
	CreatedAt int64  `json:"created_at"`
	LastUsed  int64  `json:"last_used"`
}

func toListAPIKeysResponse(k database.ListAPIKeysRow) ListAPIKeysResponse {
	var lastUsed int64 = 0

	if k.LastUsedAt.Valid {
		lastUsed = k.LastUsedAt.Time.UnixMilli()
	}
	return ListAPIKeysResponse{
		ID:        k.ID.String(),
		Name:      k.Name,
		KeySuffix: k.KeySuffix,
		CreatedAt: k.CreatedAt.Time.UnixMilli(),
		LastUsed:  lastUsed,
	}
}

func CreateAPIKey(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectIdStr := chi.URLParam(r, "project_id")
		projectId, err := utils.ParseUUID(projectIdStr)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		var req CreateAPIKeyRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			jsonError(w, "Name is required", http.StatusBadRequest)
			return
		}
		apiKey := utils.GenerateAPIKey()
		keySuffix := apiKey[len(apiKey)-16:]

		hash := sha256.Sum256([]byte(apiKey))
		hashedKey := hex.EncodeToString(hash[:])

		var key database.ApiKey
		key, err = queries.CreateAPIKey(r.Context(), database.CreateAPIKeyParams{
			ProjectID: projectId,
			Name:      req.Name,
			KeyHash:   string(hashedKey),
			KeySuffix: keySuffix,
		})
		if err != nil {
			jsonError(w, "Failed to create API key", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateAPIKeyResponse{
			APIKey:    apiKey,
			ID:        key.ID.String(),
			Name:      key.Name,
			KeySuffix: key.KeySuffix,
			CreatedAt: key.CreatedAt.Time.UnixMilli(),
			LastUsed:  key.LastUsedAt.Time.UnixMilli(),
		})
	}
}

func ListAPIKeys(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectIdStr := chi.URLParam(r, "project_id")
		projectId, err := utils.ParseUUID(projectIdStr)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		keys, err := queries.ListAPIKeys(r.Context(), projectId)
		if err != nil {
			jsonError(w, "Failed to fetch API keys", http.StatusInternalServerError)
			return
		}
		res := make([]ListAPIKeysResponse, len(keys))
		for i, k := range keys {
			res[i] = toListAPIKeysResponse(k)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func DeleteAPIKey(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyIdStr := chi.URLParam(r, "key_id")
		projectIdStr := chi.URLParam(r, "project_id")
		keyId, err := utils.ParseUUID(keyIdStr)
		if err != nil {
			jsonError(w, "Invalid API key ID", http.StatusBadRequest)
			return
		}
		projectId, err := utils.ParseUUID(projectIdStr)
		if err != nil {
			jsonError(w, "Invalid Project ID", http.StatusBadRequest)
			return
		}
		err = queries.DeleteAPIKey(r.Context(), database.DeleteAPIKeyParams{
			ID:        keyId,
			ProjectID: projectId,
		})
		if err != nil {
			jsonError(w, "Failed to delete API key", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
