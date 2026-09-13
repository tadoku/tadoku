package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

// NewHandler builds the application router without any legacy upstreams.
func NewHandler(
	application *app.Application,
	ready func(context.Context) error,
	timeout time.Duration,
	logger *slog.Logger,
) (*stdhttp.ServeMux, error) {
	if application == nil || ready == nil {
		return nil, fmt.Errorf("application and readiness are required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	mux := stdhttp.NewServeMux()
	mux.HandleFunc("GET /livez", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /readyz", withRequestTimeout(timeout, readinessHandler(ready)))
	mux.Handle(
		"GET /content/announcements/{namespace}/active",
		withRequestTimeout(timeout, listActiveAnnouncements(application, logger)),
	)

	return mux, nil
}

func withRequestTimeout(timeout time.Duration, next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func readinessHandler(ready func(context.Context) error) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := ready(r.Context()); err != nil {
			w.WriteHeader(stdhttp.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready","checks":[{"name":"postgres","status":"failed"}]}`))
			return
		}

		_, _ = w.Write([]byte(`{"status":"ready","checks":[{"name":"postgres","status":"ok"}]}`))
	})
}

func writeJSON(w stdhttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
