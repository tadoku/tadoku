package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testauth"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestApplicationStartsAndShutsDown(t *testing.T) {
	db := testpostgres.New(t)
	issuer := testauth.New(t)
	databaseURL, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	databasePort, err := strconv.Atoi(databaseURL.Port())
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()
	cfg := config{
		Port: 0, MetricsPort: 0, ServiceName: "tadoku-api-test",
		AuthzURL: upstream.URL, ContentURL: upstream.URL, ImmersionURL: upstream.URL, ProfileURL: upstream.URL,
		DialTimeout: time.Second, ResponseHeaderTimeout: time.Second, RequestTimeout: time.Second,
		IdleTimeout: time.Second, ShutdownTimeout: time.Second,
		JWKS: issuer.URL, KetoReadURL: upstream.URL, PostgresMaxConnections: 4,
		Postgres: postgresconfig.Config{Host: "127.0.0.1", Port: uint16(databasePort), Database: databaseURL.Path[1:], User: "postgres", Password: "postgres", SSLMode: "disable"},
	}
	app, err := start(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := app.wait(ctx); err != nil {
			t.Errorf("cleanup application: %v", err)
		}
	})

	_, port, err := net.SplitHostPort(app.listener.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	address := net.JoinHostPort("127.0.0.1", port)
	response, err := http.Get("http://" + address + "/readyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("got %v, want %v", response.StatusCode, http.StatusOK)
	}
	if app.pool.Config().MaxConns != 4 {
		t.Errorf("pool max=%d", app.pool.Config().MaxConns)
	}
	request, err := http.NewRequest("GET", "http://"+address+"/content/announcements/empty/active", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", issuer.Token(t, jwt.MapClaims{"sub": "guest", "iat": 1700000000}))
	active, err := (&http.Client{Timeout: time.Second}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(active.Body)
	_ = active.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if active.StatusCode != 200 || string(body) != "{\"announcements\":[]}\n" {
		t.Errorf("native listener: status=%d body=%s", active.StatusCode, body)
	}
	_, metricsPort, err := net.SplitHostPort(app.metricsListener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	metricsURL := "http://" + net.JoinHostPort("127.0.0.1", metricsPort) + "/metrics"
	metrics, err := (&http.Client{Timeout: time.Second}).Get(metricsURL)
	if err != nil {
		t.Fatal(err)
	}
	metricsBody, err := io.ReadAll(metrics.Body)
	_ = metrics.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(metricsBody), "tadoku_api_native_request_duration_seconds") {
		t.Error("native metrics not exported")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = http.Get("http://" + address + "/livez")
	if err == nil {
		t.Errorf("expected an error")
	}
	if err := app.pool.Ping(context.Background()); err == nil {
		t.Error("pool remained usable after shutdown")
	}
	if response, err := http.Get(metricsURL); err == nil {
		response.Body.Close()
		t.Error("metrics listener remained open")
	}

	// A JWKS startup failure must not leave a database pool behind.
	cfg.JWKS = upstream.URL
	if failed, err := start(cfg, slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_ = failed.wait(ctx)
		t.Error("startup accepted an unavailable JWKS")
	}
}

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_AUTHZ_URL", "http://authz")
	t.Setenv("API_CONTENT_URL", "http://content")
	t.Setenv("API_IMMERSION_URL", "http://immersion")
	t.Setenv("API_PROFILE_URL", "http://profile")
	t.Setenv("API_JWKS", "http://gateway/.well-known/jwks.json")
	t.Setenv("API_KETO_READ_URL", "http://keto-read:4466")
	for key, value := range map[string]string{"HOST": "localhost", "DATABASE": "tadoku", "USER": "tadoku", "PASSWORD": "synthetic", "SSLMODE": "disable"} {
		t.Setenv("API_POSTGRES_"+key, value)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8000 {
		t.Errorf("got %v, want %v", cfg.Port, 8000)
	}
	if cfg.MetricsPort != 9090 {
		t.Errorf("got %v, want %v", cfg.MetricsPort, 9090)
	}
	if cfg.ServiceName != "tadoku-api" {
		t.Errorf("got %v, want %v", cfg.ServiceName, "tadoku-api")
	}
	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("got %v, want %v", cfg.RequestTimeout, 30*time.Second)
	}
	if cfg.PostgresMaxConnections != 4 {
		t.Errorf("pool limit=%d want=4", cfg.PostgresMaxConnections)
	}
}
