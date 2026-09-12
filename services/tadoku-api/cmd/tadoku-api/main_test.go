package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApplicationStartsAndShutsDown(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()
	cfg := config{
		Port: 0, MetricsPort: 0, ServiceName: "tadoku-api-test",
		AuthzURL: upstream.URL, ContentURL: upstream.URL, ImmersionURL: upstream.URL, ProfileURL: upstream.URL,
		DialTimeout: time.Second, ResponseHeaderTimeout: time.Second, RequestTimeout: time.Second,
		IdleTimeout: time.Second, ShutdownTimeout: time.Second,
	}
	app, err := start(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := app.wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = http.Get("http://" + address + "/livez")
	if err == nil {
		t.Errorf("expected an error")
	}
}

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_AUTHZ_URL", "http://authz")
	t.Setenv("API_CONTENT_URL", "http://content")
	t.Setenv("API_IMMERSION_URL", "http://immersion")
	t.Setenv("API_PROFILE_URL", "http://profile")

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
}
