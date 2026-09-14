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

// Router keeps application routes behind shared middleware while allowing this
// package to attach probes and temporary legacy proxies outside it.
type Router struct {
	mux                  *stdhttp.ServeMux
	application          *stdhttp.ServeMux
	protectedApplication stdhttp.Handler
}

// Handle registers an application route behind the shared middleware.
func (r *Router) Handle(pattern string, handler stdhttp.Handler) {
	r.application.Handle(pattern, handler)
	r.mux.Handle(pattern, r.protectedApplication)
}

// HandleFunc registers an application route behind the shared middleware.
func (r *Router) HandleFunc(pattern string, handler func(stdhttp.ResponseWriter, *stdhttp.Request)) {
	r.application.HandleFunc(pattern, handler)
	r.mux.Handle(pattern, r.protectedApplication)
}

func (r *Router) ServeHTTP(w stdhttp.ResponseWriter, request *stdhttp.Request) {
	r.mux.ServeHTTP(w, request)
}

// NewHandler builds the application router without any legacy upstreams.
func NewHandler(
	application *app.Application,
	ready func(context.Context) error,
	timeout time.Duration,
	logger *slog.Logger,
	authenticate func(stdhttp.Handler) stdhttp.Handler,
	rejectBanned func(stdhttp.Handler) stdhttp.Handler,
) (*Router, error) {
	if application == nil || ready == nil {
		return nil, fmt.Errorf("application and readiness are required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("request timeout must be positive")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if authenticate == nil {
		return nil, fmt.Errorf("authentication middleware is required")
	}
	if rejectBanned == nil {
		return nil, fmt.Errorf("banned-user middleware is required")
	}

	router := &Router{
		mux:         stdhttp.NewServeMux(),
		application: stdhttp.NewServeMux(),
	}
	router.protectedApplication = withRequestTimeout(timeout, authenticate(rejectBanned(router.application)))
	router.mux.HandleFunc("GET /livez", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	router.mux.Handle("GET /readyz", withRequestTimeout(timeout, readinessHandler(ready)))
	router.Handle(
		"GET /content/announcements/{namespace}/active",
		listActiveAnnouncements(application, logger),
	)

	return router, nil
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
