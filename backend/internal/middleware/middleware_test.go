package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vknow360/otaship/backend/internal/utils"
)

func TestAdminOnly(t *testing.T) {
	// The raw token
	rawToken := "my-secret-admin-token"
	// The token the server expects (the SHA256 hex hash of the raw token)
	validTokenHash := utils.CalculateSHA256([]byte(rawToken))

	// Create the middleware
	mw := AdminOnly(validTokenHash)

	// A dummy handler that runs if auth passes
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the dummy handler
	handler := mw(nextHandler)

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
	}{
		{
			name:         "Valid Token",
			authHeader:   "Bearer my-secret-admin-token",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid Token",
			authHeader:   "Bearer wrong-token",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Missing Bearer Prefix",
			authHeader:   "my-secret-admin-token",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Empty Header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/admin/projects", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			if rr.Code == http.StatusUnauthorized {
				var resp map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("Failed to decode JSON response: %v", err)
				}
				if resp["error"] != "Unauthorized" {
					t.Errorf("Expected error 'Unauthorized', got %v", resp["error"])
				}
			}
		})
	}
}

func TestProjectKeyOnly_SkipWithoutDB(t *testing.T) {
	// ProjectKeyOnly requires a concrete *database.Queries struct.
	// Since we are not using a mock interface (emit_interface: false in sqlc),
	// testing it requires a real Postgres database connection.
	t.Skip("Skipping ProjectKeyOnly test. Requires integration test with real DB.")
}
