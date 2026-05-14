package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/utils"
)

func AdminOnly(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
				bearerToken := after
				if utils.CalculateSHA256([]byte(bearerToken)) == token {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		})
	}
}

func ProjectKeyOnly(queries *database.Queries) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
				return
			}

			if len(apiKey) < 16 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
				return
			}

			keySuffix := apiKey[len(apiKey)-16:]
			key, err := queries.GetAPIKeyBySuffix(r.Context(), keySuffix)
			if err != nil {
				slog.ErrorContext(r.Context(), "Error fetching project by key suffix",
					slog.String("suffix", keySuffix),
					slog.Any("error", err),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
				return
			}

			hash := sha256.Sum256([]byte(apiKey))
			computed := hex.EncodeToString(hash[:])
			if computed != key.KeyHash {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key"})
				return
			}
			go func() {
				queries.UpdateAPIKeyLastUsed(context.Background(), key.ID)

			}()
			next.ServeHTTP(w, r.WithContext(utils.SetProjectId(r.Context(), key.ProjectID)))
		})
	}
}
