package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/access"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

type config struct {
	Port                   int                   `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort            int                   `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName            string                `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	AuthzURL               string                `validate:"required" envconfig:"authz_url"`
	ContentURL             string                `validate:"required" envconfig:"content_url"`
	ImmersionURL           string                `validate:"required" envconfig:"immersion_url"`
	ProfileURL             string                `validate:"required" envconfig:"profile_url"`
	JWKS                   string                `validate:"required,url" envconfig:"jwks"`
	KetoReadURL            string                `validate:"required,url" envconfig:"keto_read_url"`
	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	Postgres               postgresconfig.Config `ignored:"true"`
	DialTimeout            time.Duration         `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	ResponseHeaderTimeout  time.Duration         `validate:"gt=0" envconfig:"response_header_timeout" default:"10s"`
	RequestTimeout         time.Duration         `validate:"gt=0" envconfig:"request_timeout" default:"30s"`
	IdleTimeout            time.Duration         `validate:"gt=0" envconfig:"idle_timeout" default:"30s"`
	ShutdownTimeout        time.Duration         `validate:"gt=0" envconfig:"shutdown_timeout" default:"10s"`
}

func loadConfig() (config, error) {
	cfg := config{}
	if err := envconfig.Process("API", &cfg); err != nil {
		return config{}, fmt.Errorf("load config: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate config: %w", err)
	}
	for name, raw := range map[string]string{"JWKS": cfg.JWKS, "KETO_READ_URL": cfg.KetoReadURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
			return config{}, fmt.Errorf("API_%s must be an HTTP(S) URL without credentials or fragment", name)
		}
	}
	var err error
	cfg.Postgres, err = postgresconfig.Load("API_POSTGRES", "API_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}
	return cfg, nil
}

type application struct {
	server          *http.Server
	listener        net.Listener
	serverErrors    chan error
	metricsServer   *http.Server
	metricsListener net.Listener
	shutdownTimeout time.Duration
	transport       *http.Transport
	pool            *pgxpool.Pool
	auth            *transporthttp.Authenticator
}

func start(cfg config, logger *slog.Logger) (*application, error) {
	logger = logger.With("service", cfg.ServiceName)
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: cfg.DialTimeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   25,
		IdleConnTimeout:       cfg.IdleTimeout,
		TLSHandshakeTimeout:   cfg.DialTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: time.Second,
	}
	startupContext, cancelStartup := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancelStartup()
	pool, err := postgres.Open(startupContext, cfg.Postgres.URL(), cfg.PostgresMaxConnections)
	if err != nil {
		transport.CloseIdleConnections()
		return nil, fmt.Errorf("open postgres: %s", cfg.Postgres.Redact(err))
	}
	started := false
	defer func() {
		if !started {
			pool.Close()
			transport.CloseIdleConnections()
		}
	}()
	auth, err := transporthttp.NewAuthenticator(startupContext, cfg.JWKS, &http.Client{Transport: transport, Timeout: cfg.DialTimeout})
	if err != nil {
		return nil, err
	}
	defer func() {
		if !started {
			auth.Close()
		}
	}()
	native := app.New(content.NewService(content.NewRepository(pool)), access.NewService(cfg.KetoReadURL), logger)
	handler, err := transporthttp.NewHandler(native, auth, pool.Ping, transporthttp.Upstreams{
		Authz: cfg.AuthzURL, Content: cfg.ContentURL,
		Immersion: cfg.ImmersionURL, Profile: cfg.ProfileURL,
	}, transport, cfg.RequestTimeout, metrics, logger)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("listen for API requests: %w", err)
	}
	metricsListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort))
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("start metrics server: %w", err)
	}
	metricsServer := &http.Server{
		Handler:           promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}),
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: cfg.RequestTimeout, IdleTimeout: cfg.IdleTimeout,
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout + time.Second,
		IdleTimeout:       cfg.IdleTimeout,
	}
	app := &application{
		server: server, listener: listener, serverErrors: make(chan error, 2),
		metricsServer: metricsServer, metricsListener: metricsListener, shutdownTimeout: cfg.ShutdownTimeout, transport: transport, pool: pool, auth: auth,
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.serverErrors <- fmt.Errorf("API server stopped: %w", err)
		}
	}()
	go func() {
		if err := metricsServer.Serve(metricsListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.serverErrors <- fmt.Errorf("metrics server stopped: %w", err)
		}
	}()
	started = true
	logger.Info("tadoku-api started", "address", listener.Addr(), "mode", "native-and-proxy", "postgres_max_connections", cfg.PostgresMaxConnections)
	return app, nil
}

func (app *application) wait(ctx context.Context) error {
	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-app.serverErrors:
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancel()
	shutdownErr := app.server.Shutdown(shutdownContext)
	if shutdownErr != nil {
		_ = app.server.Close()
	} // cancel remaining request contexts
	metricsErr := app.metricsServer.Shutdown(shutdownContext)
	if metricsErr != nil {
		_ = app.metricsServer.Close()
	}
	app.auth.Close()
	app.pool.Close()
	app.transport.CloseIdleConnections()
	return errors.Join(runErr, shutdownErr, metricsErr)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	app, err := start(cfg, logger)
	if err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.wait(ctx); err != nil {
		panic(err)
	}
}
