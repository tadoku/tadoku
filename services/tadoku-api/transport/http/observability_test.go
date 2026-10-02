package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
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

type requestTenantHandler struct {
	slog.Handler
	key tenant.Key
	ok  bool
}

func (handler *requestTenantHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Message == "request completed" {
		handler.key, handler.ok = tenant.FromContext(ctx)
	}
	return handler.Handler.Handle(ctx, record)
}

func TestRequestObservationUsesVerifiedTenant(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	jwks := fmt.Sprintf(`{"keys":[%s]}`, rsaJWK("observability", &privateKey.PublicKey))
	keys := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte(jwks))
	}))
	t.Cleanup(keys.Close)

	for _, test := range []struct {
		name       string
		tenant     string
		expired    bool
		omitToken  bool
		wantStatus int
		wantKind   string
		wantTenant string
	}{
		{"production", "tadoku/prod", false, false, stdhttp.StatusNoContent, "production", "tadoku/prod"},
		{"test", "e2e/observability-0123abcd", false, false, stdhttp.StatusNoContent, "test", "e2e/observability-0123abcd"},
		{"expired", "tadoku/prod", true, false, stdhttp.StatusUnauthorized, "unknown", ""},
		{"malformed tenant", "not-a-tenant", false, false, stdhttp.StatusUnauthorized, "unknown", ""},
		{"missing tenant", "", false, false, stdhttp.StatusUnauthorized, "unknown", ""},
		{"missing JWT", "", false, true, stdhttp.StatusBadRequest, "unknown", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			captured := &requestTenantHandler{Handler: slog.NewJSONHandler(&logs, nil)}
			logger := slog.New(captured)
			authenticate, err := NewJWTAuthentication(
				t.Context(), keys.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, logger,
			)
			if err != nil {
				t.Fatal(err)
			}
			passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
			registry := prometheus.NewRegistry()
			router, err := NewHandler(
				app.New(app.Dependencies{}),
				func(context.Context) error { return nil },
				time.Second,
				registry,
				logger,
				authenticate,
				passthrough,
				passthrough,
			)
			if err != nil {
				t.Fatal(err)
			}
			router.HandleFunc("GET /observation", func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				w.WriteHeader(stdhttp.StatusNoContent)
			})

			now := time.Now()
			expiresAt := now.Add(time.Hour)
			if test.expired {
				expiresAt = now.Add(-time.Minute)
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, &userClaims{
				Tenant: test.tenant,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "guest",
					IssuedAt:  jwt.NewNumericDate(now.Add(-time.Hour)),
					ExpiresAt: jwt.NewNumericDate(expiresAt),
				},
			})
			token.Header["kid"] = "observability"
			signed, err := token.SignedString(privateKey)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(stdhttp.MethodGet, "/observation", nil)
			if !test.omitToken {
				request.Header.Set("Authorization", "Bearer "+signed)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status=%d, want %d", response.Code, test.wantStatus)
			}
			if captured.key.String() != test.wantTenant || captured.ok != (test.wantTenant != "") {
				t.Errorf("completion context tenant=%q known=%t, want %q", captured.key.String(), captured.ok, test.wantTenant)
			}
			if test.wantTenant == "" {
				var event map[string]any
				if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
					t.Fatal(err)
				}
				if event["msg"] != "request completed" || event["tenant"] != "unknown" {
					t.Errorf("pre-auth completion event=%v, want tenant=unknown", event)
				}
				if count := bytes.Count(logs.Bytes(), []byte(`"tenant"`)); count != 1 {
					t.Errorf("pre-auth completion has %d tenant attributes, want exactly one", count)
				}
			}
			families, err := registry.Gather()
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, family := range families {
				if family.GetName() != "tadoku_api_proxy_request_duration_seconds" {
					continue
				}
				for _, metric := range family.GetMetric() {
					labels := make(map[string]string)
					for _, label := range metric.GetLabel() {
						labels[label.GetName()] = label.GetValue()
					}
					if labels["route"] != "GET /observation" {
						continue
					}
					found = true
					if labels["tenant_kind"] != test.wantKind || len(labels) != 5 {
						t.Errorf("histogram labels=%v, want bounded tenant_kind=%q and existing four labels", labels, test.wantKind)
					}
					if metric.GetHistogram().GetSampleCount() != 1 {
						t.Errorf("histogram sample count=%d, want 1", metric.GetHistogram().GetSampleCount())
					}
				}
			}
			if !found {
				t.Error("request histogram sample not found")
			}
		})
	}
}

func TestRequestMetricUsesMatchedPattern(t *testing.T) {
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
		app.New(app.Dependencies{}),
		func(context.Context) error { return nil },
		time.Second,
		registry,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		authenticate,
		passthrough,
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
		t.Errorf("response changed with correlation ID %q", got)
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
	t.Fatal("request histogram sample not found")
}
