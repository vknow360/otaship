package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func HealthCheck(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbErr := db.Ping(r.Context())

		status := "ok"
		httpStatus := http.StatusOK

		if dbErr != nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}

		result := map[string]interface{}{
			"status": status,
			"db":     statusString(dbErr),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		json.NewEncoder(w).Encode(result)
	}
}

func statusString(err error) string {
	if err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}
