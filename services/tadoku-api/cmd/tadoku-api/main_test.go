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

	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestApplicationStartsAndShutsDown(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	databaseURL, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}

	databasePort, err := strconv.Atoi(databaseURL.Port())
	if err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(jwks.Close)

	cfg := config{
		Port:        0,
		MetricsPort: 0,
		ServiceName: "tadoku-api-test",
		JWKS:        jwks.URL,
		KetoReadURL: upstream.URL,

		AuthzURL:     upstream.URL,
		ContentURL:   upstream.URL,
		ImmersionURL: upstream.URL,
		ProfileURL:   upstream.URL,

		DialTimeout:           time.Second,
		ResponseHeaderTimeout: time.Second,
		RequestTimeout:        time.Second,
		IdleTimeout:           time.Second,
		ShutdownTimeout:       time.Second,

		PostgresMaxConnections: 4,
		Postgres: postgresconfig.Config{
			Host:     "127.0.0.1",
			Port:     uint16(databasePort),
			Database: databaseURL.Path[1:],
			User:     "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		},
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

	// Verify the application is ready to accept requests.
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

	// Existing process metrics still use their own listener.
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
	if !strings.Contains(string(metricsBody), "go_goroutines") {
		t.Error("process metrics not exported")
	}

	// Shutdown closes both listeners and the shared database pool.
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
}

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_JWKS", "http://jwks.test")
	t.Setenv("API_KETO_READ_URL", "http://keto-read.test")
	t.Setenv("API_AUTHZ_URL", "http://authz")
	t.Setenv("API_CONTENT_URL", "http://content")
	t.Setenv("API_IMMERSION_URL", "http://immersion")
	t.Setenv("API_PROFILE_URL", "http://profile")
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
	if cfg.JWKS != "http://jwks.test" {
		t.Errorf("JWKS=%q", cfg.JWKS)
	}
	if cfg.KetoReadURL != "http://keto-read.test" {
		t.Errorf("Keto read URL=%q", cfg.KetoReadURL)
	}
	t.Setenv("API_JWKS", "")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "JWKS") {
		t.Errorf("missing JWKS configuration error=%v", err)
	}
	t.Setenv("API_JWKS", "http://jwks.test")
	for _, ketoURL := range []string{"", "not-a-url", "ftp://keto-read.test"} {
		t.Setenv("API_KETO_READ_URL", ketoURL)
		if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "KetoReadURL") {
			t.Errorf("Keto read URL %q error=%v", ketoURL, err)
		}
	}
}

func TestApplicationRejectsUnavailableJWKSBeforeStarting(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(provider.Close)

	app, err := start(config{
		JWKS:        provider.URL,
		DialTimeout: time.Second,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if app != nil || err == nil || !strings.Contains(err.Error(), "fetch authentication JWKS") {
		t.Errorf("startup with unavailable JWKS: application=%v error=%v", app, err)
	}
}
