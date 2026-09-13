package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

const activeAnnouncementsRoute = "/content/announcements/{namespace}/active"

func NewHandler(application *app.Application, auth *Authenticator, ready func(context.Context) error, upstreams Upstreams, transport stdhttp.RoundTripper, timeout time.Duration, registerer prometheus.Registerer, logger *slog.Logger) (stdhttp.Handler, error) {
	if application == nil || auth == nil || ready == nil {
		return nil, fmt.Errorf("native application, authentication and readiness are required")
	}
	fallback, err := NewProxyHandler(upstreams, transport, timeout, registerer, logger)
	if err != nil {
		return nil, err
	}
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tadoku_api_native_request_duration_seconds", Help: "Duration of native API requests.",
	}, []string{"method", "route", "mode", "status"})
	if err := registerer.Register(duration); err != nil {
		return nil, fmt.Errorf("register native metrics: %w", err)
	}
	mux := stdhttp.NewServeMux()
	mux.Handle("/", fallback)
	mux.HandleFunc("GET /readyz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
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
	// This static registration owns GET only. In particular, ServeMux's automatic
	// GET -> HEAD matching must not silently migrate HEAD or OPTIONS from Echo.
	mux.HandleFunc(activeAnnouncementsRoute, func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodGet {
			fallback.ServeHTTP(w, r)
			return
		}
		started := time.Now() // elapsed time is deliberately not the business clock
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(withCorrelationID(ctx, r.Header.Get(correlationHeader)))
		w.Header().Set(correlationHeader, correlationID(r))
		recorder := &statusRecorder{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(ctx, "native request panicked", "route", activeAnnouncementsRoute)
				writeJSON(recorder, 500, map[string]string{"message": "Internal Server Error"})
			}
			status := recorder.status
			if status == 0 {
				status = 200
			}
			elapsed := time.Since(started)
			duration.WithLabelValues("GET", activeAnnouncementsRoute, "native", strconv.Itoa(status)).Observe(elapsed.Seconds())
			logger.InfoContext(ctx, "request completed", "correlation_id", correlationID(r), "method", "GET", "route", activeAnnouncementsRoute, "mode", "native", "status", status, "latency", elapsed)
		}()
		principal, status := auth.authenticate(r)
		if status != 0 {
			message := "missing or malformed jwt"
			if status == stdhttp.StatusUnauthorized {
				message = "invalid or expired jwt"
			}
			if status == stdhttp.StatusInternalServerError {
				message = "Internal Server Error"
			}
			writeJSON(recorder, status, map[string]string{"message": message})
			return
		}
		items, err := application.ActiveAnnouncements(r.Context(), principal, r.PathValue("namespace"))
		if err != nil {
			status := stdhttp.StatusInternalServerError
			if errors.Is(err, app.ErrForbidden) {
				status = stdhttp.StatusForbidden
			} else {
				logger.ErrorContext(ctx, "active announcements failed", "correlation_id", correlationID(r), "error", err)
			}
			recorder.WriteHeader(status)
			return // claimed native errors never fall through to the proxy
		}
		response := openapi.ContentAnnouncements{Announcements: make([]openapi.ContentAnnouncement, 0, len(items))}
		for _, item := range items {
			response.Announcements = append(response.Announcements, openapi.ContentAnnouncement{
				Id: &item.ID, Namespace: &item.Namespace, Title: item.Title, Content: item.Content,
				Style: openapi.ContentAnnouncementStyle(item.Style), Href: item.Href,
				StartsAt: item.StartsAt, EndsAt: item.EndsAt, CreatedAt: &item.CreatedAt, UpdatedAt: &item.UpdatedAt,
			})
		}
		writeJSON(recorder, stdhttp.StatusOK, response)
	})
	return mux, nil
}

func writeJSON(w stdhttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
