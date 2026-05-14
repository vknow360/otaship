package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func InitLogger(level, format string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slogLevel(level),
	}

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	handler = &contextHandler{handler}
	slog.SetDefault(slog.New(handler))
}

type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		sw := &statusWriter{ResponseWriter: w, status: 200}

		slog.InfoContext(ctx, "Request started",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		next.ServeHTTP(sw, r.WithContext(ctx))

		slog.InfoContext(ctx, "Request completed",
			slog.Int("status", sw.status),
			slog.Duration("latency", time.Since(start)),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
	})
}

func slogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
