package main

import (
	"context"
	"crypto/rand"
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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestApplicationStartsAndShutsDown(t *testing.T) {
	cfg := validApplicationConfig(t)
	observer, err := valkeygo.NewClient(valkeygo.MustParseURL(cfg.ValkeyURL))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(observer.Close)

	ctx, cancel := context.WithCancel(t.Context())
	application, err := start(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stopped := false
	t.Cleanup(func() {
		if stopped {
			return
		}
		cancel()
		if err := application.wait(); err != nil {
			t.Errorf("cleanup application: %v", err)
		}
	})

	clientID, err := application.valkey.Do(t.Context(), application.valkey.B().ClientId().Build()).AsInt64()
	if err != nil {
		t.Fatalf("owned Valkey client ID: %v", err)
	}

	// Verify the application is ready to accept requests.
	_, port, err := net.SplitHostPort(application.listener.Addr().String())
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
	if application.pool.Config().MaxConns != 4 {
		t.Errorf("pool max=%d", application.pool.Config().MaxConns)
	}
	if got := application.pool.Config().ConnConfig.RuntimeParams["application_name"]; got != cfg.ServiceName {
		t.Errorf("application_name=%q want=%q", got, cfg.ServiceName)
	}

	// Existing process metrics still use their own listener.
	_, metricsPort, err := net.SplitHostPort(application.metricsListener.Addr().String())
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

	// Shutdown closes both listeners and the owned database and Valkey clients.
	cancel()
	if err := application.wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stopped = true

	_, err = http.Get("http://" + address + "/livez")
	if err == nil {
		t.Errorf("expected an error")
	}
	if err := application.pool.Ping(context.Background()); err == nil {
		t.Error("pool remained usable after shutdown")
	}
	if response, err := http.Get(metricsURL); err == nil {
		response.Body.Close()
		t.Error("metrics listener remained open")
	}
	if err := application.valkey.Do(t.Context(), application.valkey.B().Ping().Build()).Error(); !errors.Is(err, valkeygo.ErrClosing) {
		t.Errorf("closed Valkey client PING error=%v, want ErrClosing", err)
	}
	waitForValkeyDisconnect(t, observer, "id="+strconv.FormatInt(clientID, 10))
}

func validApplicationConfig(t *testing.T) config {
	t.Helper()
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
	valkeyURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}

	return config{
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
		ValkeyURL:     valkeyURL,
		ValkeyTimeout: time.Second,
	}
}

func TestLoadConfigUsesValidatedDefaults(t *testing.T) {
	t.Setenv("API_JWKS", "http://jwks.test")
	t.Setenv("API_KETO_READ_URL", "http://keto-read.test")
	t.Setenv("API_AUTHZ_URL", "http://authz")
	t.Setenv("API_CONTENT_URL", "http://content")
	t.Setenv("API_IMMERSION_URL", "http://immersion")
	t.Setenv("API_PROFILE_URL", "http://profile")
	t.Setenv("API_VALKEY_URL", "redis://valkey:6379")
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
	if cfg.ValkeyURL != "redis://valkey:6379" || cfg.ValkeyTimeout != time.Second {
		t.Errorf("Valkey URL=%q timeout=%v", cfg.ValkeyURL, cfg.ValkeyTimeout)
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
	t.Setenv("API_MAX_TOKEN_AGE", "24h")
	t.Setenv("API_VALKEY_TIMEOUT", "0s")
	if _, err := loadConfig(); err == nil || !strings.Contains(err.Error(), "ValkeyTimeout") {
		t.Errorf("invalid Valkey timeout error=%v", err)
	}
}

func TestApplicationReadinessDoesNotDependOnValkey(t *testing.T) {
	cfg := validApplicationConfig(t)
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reserved.Addr().String()
	_ = reserved.Close()
	cfg.ValkeyURL = "redis://" + address
	cfg.ValkeyTimeout = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(t.Context())
	application, err := start(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("degraded startup: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		if err := application.wait(); err != nil {
			t.Error(err)
		}
	})

	_, port, err := net.SplitHostPort(application.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Get("http://" + net.JoinHostPort("127.0.0.1", port) + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("readiness status=%d", response.StatusCode)
	}
}

func TestApplicationClosesValkeyWhenLaterStartupFails(t *testing.T) {
	cfg := validApplicationConfig(t)
	observer, err := valkeygo.NewClient(valkeygo.MustParseURL(cfg.ValkeyURL))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(observer.Close)
	username := "tadoku_api_test_" + rand.Text()
	if err := observer.Do(t.Context(), observer.B().AclSetuser().Username(username).Rule("reset", "on", ">synthetic", "~*", "+@all").Build()).Error(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := observer.Do(ctx, observer.B().AclDeluser().Username(username).Build()).Error(); err != nil {
			t.Errorf("delete test Valkey user: %v", err)
		}
	})
	valkeyURL, err := url.Parse(cfg.ValkeyURL)
	if err != nil {
		t.Fatal(err)
	}
	valkeyURL.User = url.UserPassword(username, "synthetic")
	cfg.ValkeyURL = valkeyURL.String()

	occupied, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = occupied.Close() })
	_, port, err := net.SplitHostPort(occupied.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	cfg.Port, err = strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	application, err := start(t.Context(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if application != nil || err == nil || !strings.Contains(err.Error(), "listen for API requests") {
		t.Fatalf("application=%v error=%v", application, err)
	}

	waitForValkeyDisconnect(t, observer, "user="+username)
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

// Close completes locally before Valkey necessarily processes the disconnect.
func waitForValkeyDisconnect(t *testing.T, observer valkeygo.Client, clientField string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	var remaining []string
	for {
		clients, err := observer.Do(ctx, observer.B().ClientList().Build()).ToString()
		if err != nil {
			t.Fatalf("observe Valkey disconnect for %s: %v; last matching clients: %s", clientField, err, strings.Join(remaining, "\n"))
		}
		remaining = remaining[:0]
		for _, line := range strings.Split(clients, "\n") {
			if strings.Contains(line, clientField+" ") {
				remaining = append(remaining, line)
			}
		}
		if len(remaining) == 0 {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("Valkey clients remained open for %s: %s", clientField, strings.Join(remaining, "\n"))
		case <-ticker.C:
		}
	}
}
