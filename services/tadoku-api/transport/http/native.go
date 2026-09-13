package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

const activeAnnouncementsRoute = "/content/announcements/{namespace}/active"

func NewHandler(
	application *app.Application,
	ready func(context.Context) error,
	upstreams Upstreams,
	transport stdhttp.RoundTripper,
	timeout time.Duration,
	registerer prometheus.Registerer,
	logger *slog.Logger,
) (stdhttp.Handler, error) {
	if application == nil || ready == nil {
		return nil, fmt.Errorf("native application and readiness are required")
	}

	fallback, err := NewProxyHandler(upstreams, transport, timeout, registerer, logger)
	if err != nil {
		return nil, err
	}

	announcements := listActiveAnnouncements(application, logger)

	mux := stdhttp.NewServeMux()
	mux.Handle("/", fallback)
	mux.Handle("GET /readyz", readinessHandler(ready, timeout))
	mux.HandleFunc(activeAnnouncementsRoute, func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		// Claim GET only; the remaining methods still belong to the proxy.
		if r.Method != stdhttp.MethodGet {
			fallback.ServeHTTP(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		announcements.ServeHTTP(w, r.WithContext(ctx))
	})

	return mux, nil
}

func readinessHandler(ready func(context.Context) error, timeout time.Duration) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")
		if err := ready(ctx); err != nil {
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
