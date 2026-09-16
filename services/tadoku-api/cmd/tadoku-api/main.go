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
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

type config struct {
	Port        int    `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort int    `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName string `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	JWKS        string `validate:"required"`
	KetoReadURL string `validate:"required" envconfig:"keto_read_url"`

	AuthzURL     string `validate:"required" envconfig:"authz_url"`
	ContentURL   string `validate:"required" envconfig:"content_url"`
	ImmersionURL string `validate:"required" envconfig:"immersion_url"`
	ProfileURL   string `validate:"required" envconfig:"profile_url"`

	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	Postgres               postgresconfig.Config `ignored:"true"`

	DialTimeout           time.Duration `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	ResponseHeaderTimeout time.Duration `validate:"gt=0" envconfig:"response_header_timeout" default:"10s"`
	RequestTimeout        time.Duration `validate:"gt=0" envconfig:"request_timeout" default:"30s"`
	IdleTimeout           time.Duration `validate:"gt=0" envconfig:"idle_timeout" default:"30s"`
	ShutdownTimeout       time.Duration `validate:"gt=0" envconfig:"shutdown_timeout" default:"10s"`
}

func loadConfig() (config, error) {
	cfg := config{}
	if err := envconfig.Process("API", &cfg); err != nil {
		return config{}, fmt.Errorf("load config: %w", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate config: %w", err)
	}
	ketoURL, err := url.ParseRequestURI(cfg.KetoReadURL)
	if err != nil || ketoURL.Host == "" || (ketoURL.Scheme != "http" && ketoURL.Scheme != "https") {
		return config{}, fmt.Errorf("validate config: KetoReadURL must be an HTTP(S) URL")
	}

	cfg.Postgres, err = postgresconfig.Load("API_POSTGRES", "API_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

type application struct {
	server       *http.Server
	listener     net.Listener
	serverErrors chan error

	metricsServer   *http.Server
	metricsListener net.Listener

	shutdownTimeout time.Duration
	transport       *http.Transport
	pool            *pgxpool.Pool
}

func start(cfg config, logger *slog.Logger) (*application, error) {
	logger = logger.With("service", cfg.ServiceName)
	authenticate, err := transporthttp.NewJWTAuthentication(cfg.JWKS, cfg.DialTimeout)
	if err != nil {
		return nil, err
	}

	metrics := prometheus.NewRegistry()
	metrics.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

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

	pool, err := postgres.Open(startupContext, cfg.Postgres.WithApplicationName(cfg.ServiceName).URL(), cfg.PostgresMaxConnections)
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

	keto := ketoclient.NewReadClient(cfg.KetoReadURL)
	permissionChecker := permissions.NewKetoChecker(keto)
	contentRepository := content.NewAnnouncementsRepository(pool)
	contentService := content.NewService(contentRepository)
	api := app.New(contentService, pool, permissionChecker)
	rejectBanned := newBannedUserMiddleware(cfg.KetoReadURL, logger)

	handler, err := transporthttp.NewHandler(api, pool.Ping, cfg.RequestTimeout, logger, authenticate, rejectBanned)
	if err != nil {
		return nil, err
	}

	// Temporary legacy routes; the application router stands on its own.
	upstreams := transporthttp.Upstreams{
		Authz:     cfg.AuthzURL,
		Content:   cfg.ContentURL,
		Immersion: cfg.ImmersionURL,
		Profile:   cfg.ProfileURL,
	}
	err = transporthttp.RegisterProxyRoutes(
		handler,
		upstreams,
		transport,
		cfg.RequestTimeout,
		metrics,
		logger,
	)
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
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout + time.Second,
		IdleTimeout:       cfg.IdleTimeout,
	}

	app := &application{
		server:          server,
		listener:        listener,
		serverErrors:    make(chan error, 2),
		metricsServer:   metricsServer,
		metricsListener: metricsListener,
		shutdownTimeout: cfg.ShutdownTimeout,
		transport:       transport,
		pool:            pool,
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
	logger.Info("tadoku-api started",
		"address", listener.Addr(),
		"postgres_max_connections", cfg.PostgresMaxConnections,
	)

	return app, nil
}

func newBannedUserMiddleware(ketoReadURL string, logger *slog.Logger) func(http.Handler) http.Handler {
	keto := ketoclient.NewReadClient(ketoReadURL)
	return transporthttp.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return keto.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
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
		// Cancel request contexts that outlasted graceful shutdown.
		_ = app.server.Close()
	}

	metricsErr := app.metricsServer.Shutdown(shutdownContext)
	if metricsErr != nil {
		_ = app.metricsServer.Close()
	}

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
