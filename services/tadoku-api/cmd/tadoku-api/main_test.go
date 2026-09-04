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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	_, port, err := net.SplitHostPort(app.listener.Addr().String())
	require.NoError(t, err)
	address := net.JoinHostPort("127.0.0.1", port)
	response, err := http.Get("http://" + address + "/readyz")
	require.NoError(t, err)
	defer response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, app.wait(ctx))

	_, err = http.Get("http://" + address + "/livez")
	assert.Error(t, err)
}

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_AUTHZ_URL", "http://authz")
	t.Setenv("API_CONTENT_URL", "http://content")
	t.Setenv("API_IMMERSION_URL", "http://immersion")
	t.Setenv("API_PROFILE_URL", "http://profile")

	cfg, err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8000, cfg.Port)
	assert.Equal(t, 9090, cfg.MetricsPort)
	assert.Equal(t, "tadoku-api", cfg.ServiceName)
	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
}
