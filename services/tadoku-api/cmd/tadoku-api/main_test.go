package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
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
		MaxTokenAge:           24 * time.Hour,
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

	ctx, cancel := context.WithCancel(t.Context())
	app, err := start(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		if err := app.wait(); err != nil {
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
	if got := app.pool.Config().ConnConfig.RuntimeParams["application_name"]; got != cfg.ServiceName {
		t.Errorf("application_name=%q want=%q", got, cfg.ServiceName)
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
	for _, name := range []string{
		"tadoku_api_postgres_pool_acquire_count_total",
		"tadoku_api_postgres_pool_acquired_connections",
		"tadoku_api_postgres_pool_empty_acquire_count_total",
		"tadoku_api_postgres_pool_acquire_duration_seconds_total",
	} {
		if !strings.Contains(string(metricsBody), name) {
			t.Errorf("pool metric %q not exported", name)
		}
	}

	// Shutdown closes both listeners and the shared database pool.
	cancel()
	if err := app.wait(); err != nil {
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
	if cfg.MaxTokenAge != 24*time.Hour {
		t.Errorf("maximum token age=%v, want %v", cfg.MaxTokenAge, 24*time.Hour)
	}
	if cfg.JWTIssuer != "" {
		t.Errorf("JWT issuer=%q, want empty", cfg.JWTIssuer)
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
	t.Setenv("API_KETO_READ_URL", "http://keto-read.test")
	t.Setenv("API_JWT_ISSUER", "https://issuer.example.test/")
	issuerConfig, err := loadConfig()
	if err != nil {
		t.Fatalf("load issuer configuration: %v", err)
	}
	if issuerConfig.JWTIssuer != "https://issuer.example.test/" {
		t.Errorf("JWT issuer=%q", issuerConfig.JWTIssuer)
	}
	t.Setenv("API_MAX_TOKEN_AGE", "0s")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "MaxTokenAge") {
		t.Errorf("invalid maximum token age error=%v", err)
	}
}

func TestApplicationRejectsUnavailableJWKSBeforeStarting(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(provider.Close)

	app, err := start(t.Context(), config{
		JWKS:        provider.URL,
		DialTimeout: time.Second,
		MaxTokenAge: 24 * time.Hour,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if app != nil || err == nil || !strings.Contains(err.Error(), "fetch authentication JWKS") {
		t.Errorf("startup with unavailable JWKS: application=%v error=%v", app, err)
	}
}

func TestApplicationCancelsStalledJWKSFetch(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
		close(requestCanceled)
	}))
	t.Cleanup(provider.Close)

	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() {
		_, err := start(ctx, config{
			JWKS:        provider.URL,
			DialTimeout: time.Second,
			MaxTokenAge: 24 * time.Hour,
		}, slog.New(slog.NewTextHandler(io.Discard, nil)))
		result <- err
	}()

	<-requestStarted
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("canceled startup error=%v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("startup did not return promptly after cancellation")
		<-result
	}
	select {
	case <-requestCanceled:
	case <-time.After(200 * time.Millisecond):
		t.Error("JWKS provider did not observe request cancellation")
	}
}

func TestMainLogsConfigErrorsAndExitsOne(t *testing.T) {
	if os.Getenv("TADOKU_API_MAIN_HELPER") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainLogsConfigErrorsAndExitsOne$")
	cmd.Env = []string{"TADOKU_API_MAIN_HELPER=1", "API_JWKS="}
	output, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("main error=%v output=%s", err, output)
	}

	log := string(output)
	if !strings.Contains(log, `"level":"ERROR"`) || !strings.Contains(log, `"msg":"load configuration"`) {
		t.Errorf("main output=%s", output)
	}
	if strings.Contains(log, "panic:") {
		t.Errorf("main output contains panic: %s", output)
	}
}
