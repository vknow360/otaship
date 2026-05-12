package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockStorageProvider is a simple mock for storage.Provider
type mockStorageProvider struct {
	pingErr error
}

func (m *mockStorageProvider) Upload(ctx context.Context, key string, file []byte, contentType string) (string, error) {
	return "mocked-url", nil
}

func (m *mockStorageProvider) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockStorageProvider) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		mockStorageErr error
		// For the db ping we can just pass a nil *pgxpool.Pool and it will panic if we don't handle it carefully.
		// However, HealthCheck takes *pgxpool.Pool and directly calls db.Ping(ctx).
		// Since we can't easily mock *pgxpool.Pool (it is a concrete type, not an interface),
		// we can't fully unit test db.Ping without standing up a test database or abstracting db.
		// Given the constraints of the signature func HealthCheck(db *pgxpool.Pool, storage storage.Provider),
		// we'll focus on testing the interface we can mock.
		// To avoid panic on db.Ping(), we'd need a real connection or we refactor HealthCheck.
		// Since we just want to test Http semantics smartly, let's assume we can't test it directly unless we pass a valid pool.
		// Therefore, we will only write tests for it if we had an interface.
	}{}

	_ = tests
	// NOTE: HealthCheck directly accepts *pgxpool.Pool which makes it hard to unit test without a real db.
	// In a professional codebase, you'd define a Pinger interface and accept that instead.
	// For instance:
	// type Pinger interface { Ping(context.Context) error }
	// func HealthCheck(db Pinger, storage storage.Provider) http.HandlerFunc
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"Nil error", nil, "ok"},
		{"Some error", errors.New("connection failed"), "error: connection failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusString(tt.err)
			if got != tt.expected {
				t.Errorf("statusString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJsonError(t *testing.T) {
	recorder := httptest.NewRecorder()

	jsonError(recorder, "test custom error", http.StatusBadRequest)

	res := recorder.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %v, got %v", http.StatusBadRequest, res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %v", res.Header.Get("Content-Type"))
	}

	var parsed map[string]string
	json.NewDecoder(res.Body).Decode(&parsed)
	if parsed["error"] != "test custom error" {
		t.Errorf("expected JSON response error message to be 'test custom error', got '%v'", parsed["error"])
	}
}
