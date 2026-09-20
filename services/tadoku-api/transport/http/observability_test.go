package http

import (
	"context"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

func TestGeneratedCorrelationIDIsUUIDv7(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	request = request.WithContext(withCorrelationID(request.Context(), ""))

	generated := correlationID(request)
	id, err := uuid.Parse(generated)
	if err != nil {
		t.Fatalf("parse generated correlation ID: %v", err)
	}
	if got := id.String(); generated != got {
		t.Errorf("generated correlation ID = %q, want canonical %q", generated, got)
	}
	if got := id.Version(); got != uuid.Version(7) {
		t.Errorf("version = %d, want 7", got)
	}
}

func TestNativeRequestMetricUsesMatchedPattern(t *testing.T) {
	registry := prometheus.NewRegistry()
	var downstreamCorrelationID string
	authenticate := func(stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			downstreamCorrelationID = r.Header.Get(correlationHeader)
			w.WriteHeader(stdhttp.StatusUnauthorized)
		})
	}
	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	router, err := NewHandler(
		app.New(nil, nil, nil, nil, nil, nil, nil, nil),
		func(context.Context) error { return nil },
		time.Second,
		registry,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		authenticate,
		passthrough,
	)
	if err != nil {
		t.Fatal(err)
	}

	requestPath := "/content/announcements/main/11111111-1111-4111-8111-111111111111"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, requestPath, nil))

	if response.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", response.Code, stdhttp.StatusUnauthorized)
	}
	if downstreamCorrelationID == "" {
		t.Error("downstream request has no correlation ID")
	}
	if got := response.Header().Get(correlationHeader); got != "" {
		t.Errorf("native response changed with correlation ID %q", got)
	}
	wantLabels := map[string]string{
		"mode":     "native",
		"route":    "GET /content/announcements/{namespace}/{id}",
		"status":   "401",
		"upstream": "",
	}
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range metricFamilies {
		if family.GetName() != "tadoku_api_proxy_request_duration_seconds" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["mode"] != wantLabels["mode"] {
				continue
			}
			for name, want := range wantLabels {
				if got := labels[name]; got != want {
					t.Errorf("label %s=%q, want %q", name, got, want)
				}
			}
			if labels["route"] == requestPath {
				t.Errorf("route label contains request path %q", requestPath)
			}
			if got := metric.GetHistogram().GetSampleCount(); got != 1 {
				t.Errorf("sample count=%d, want 1", got)
			}
			return
		}
	}
	t.Fatal("native request histogram sample not found")
}
