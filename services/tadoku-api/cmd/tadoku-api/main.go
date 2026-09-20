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
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	kratosapi "github.com/ory/kratos-client-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	valkeyinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
	valkeygo "github.com/valkey-io/valkey-go"
)

type config struct {
	Port             int           `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort      int           `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName      string        `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	JWKS             string        `validate:"required"`
	JWTIssuer        string        `envconfig:"jwt_issuer"`
	KetoReadURL      string        `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL     string        `validate:"required" envconfig:"keto_write_url"`
	KetoWriteTimeout time.Duration `validate:"gt=0" envconfig:"keto_write_timeout" default:"2s"`

	KratosAdminURL string        `validate:"required" envconfig:"kratos_admin_url"`
	KratosTimeout  time.Duration `validate:"gt=0" envconfig:"kratos_timeout" default:"2s"`

	AuthzURL     string `validate:"required" envconfig:"authz_url"`
	ContentURL   string `validate:"required" envconfig:"content_url"`
	ImmersionURL string `validate:"required" envconfig:"immersion_url"`
	ProfileURL   string `validate:"required" envconfig:"profile_url"`

	PostgresMaxConnections int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	Postgres               postgresconfig.Config `ignored:"true"`
	ValkeyURL              string                `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout          time.Duration         `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`

	DialTimeout           time.Duration `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	MaxTokenAge           time.Duration `validate:"gt=0" envconfig:"max_token_age" default:"24h"`
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
	ketoWriteURL, err := url.Parse(cfg.KetoWriteURL)
	if err != nil || ketoWriteURL.Hostname() == "" || (ketoWriteURL.Scheme != "http" && ketoWriteURL.Scheme != "https") ||
		ketoWriteURL.User != nil || ketoWriteURL.RawQuery != "" || ketoWriteURL.ForceQuery || strings.Contains(cfg.KetoWriteURL, "#") {
		return config{}, fmt.Errorf("validate config: KetoWriteURL must be an HTTP(S) URL without credentials, query or fragment")
	}
	cfg.KetoWriteURL = strings.TrimRight(cfg.KetoWriteURL, "/")
	kratosURL, err := url.Parse(cfg.KratosAdminURL)
	if err != nil || kratosURL.Hostname() == "" || (kratosURL.Scheme != "http" && kratosURL.Scheme != "https") ||
		kratosURL.User != nil || kratosURL.RawQuery != "" || kratosURL.ForceQuery || strings.Contains(cfg.KratosAdminURL, "#") {
		return config{}, fmt.Errorf("validate config: KratosAdminURL must be an HTTP(S) URL without credentials, query or fragment")
	}
	cfg.KratosAdminURL = strings.TrimRight(cfg.KratosAdminURL, "/")

	cfg.Postgres, err = postgresconfig.Load("API_POSTGRES", "API_POSTGRES_URL")
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

type application struct {
	ctx          context.Context
	cancel       context.CancelFunc
	server       *http.Server
	listener     net.Listener
	serverErrors chan error

	metricsServer   *http.Server
	metricsListener net.Listener

	shutdownTimeout time.Duration
	transport       *http.Transport
	pool            *pgxpool.Pool
	valkey          valkeygo.Client
	kratos          *kratosapi.APIClient
	keto            *ketoclient.Client
	userCacheDone   chan struct{}
}

type pgxPoolCollector struct {
	pool                *pgxpool.Pool
	acquireCount        *prometheus.Desc
	acquiredConnections *prometheus.Desc
	emptyAcquireCount   *prometheus.Desc
	acquireDuration     *prometheus.Desc
}

func newPGXPoolCollector(pool *pgxpool.Pool) *pgxPoolCollector {
	return &pgxPoolCollector{
		pool: pool,
		acquireCount: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquire_count_total",
			"Total number of successful PostgreSQL pool acquisitions.",
			nil,
			nil,
		),
		acquiredConnections: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquired_connections",
			"Number of PostgreSQL connections currently acquired from the pool.",
			nil,
			nil,
		),
		emptyAcquireCount: prometheus.NewDesc(
			"tadoku_api_postgres_pool_empty_acquire_count_total",
			"Total number of successful PostgreSQL pool acquisitions that waited for a connection.",
			nil,
			nil,
		),
		acquireDuration: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquire_duration_seconds_total",
			"Total time spent on successful PostgreSQL pool acquisitions.",
			nil,
			nil,
		),
	}
}

func (c *pgxPoolCollector) Describe(descriptions chan<- *prometheus.Desc) {
	descriptions <- c.acquireCount
	descriptions <- c.acquiredConnections
	descriptions <- c.emptyAcquireCount
	descriptions <- c.acquireDuration
}

func (c *pgxPoolCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := c.pool.Stat()
	metrics <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(stats.AcquireCount()))
	metrics <- prometheus.MustNewConstMetric(c.acquiredConnections, prometheus.GaugeValue, float64(stats.AcquiredConns()))
	metrics <- prometheus.MustNewConstMetric(c.emptyAcquireCount, prometheus.CounterValue, float64(stats.EmptyAcquireCount()))
	metrics <- prometheus.MustNewConstMetric(c.acquireDuration, prometheus.CounterValue, stats.AcquireDuration().Seconds())
}

func start(ctx context.Context, cfg config, logger *slog.Logger) (*application, error) {
	ctx, cancel := context.WithCancel(ctx)
	started := false
	defer func() {
		if !started {
			cancel()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
	}

	logger = logger.With("service", cfg.ServiceName)
	authenticate, err := transporthttp.NewJWTAuthentication(ctx, cfg.JWKS, cfg.DialTimeout, cfg.MaxTokenAge, cfg.JWTIssuer, logger)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
	}

	metrics := prometheus.NewRegistry()

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
	defer func() {
		if !started {
			transport.CloseIdleConnections()
		}
	}()

	kratosHTTP := &http.Client{
		Transport: transport,
		Timeout:   cfg.KratosTimeout,
	}
	kratos := kratosclient.NewAPIClient(cfg.KratosAdminURL, kratosclient.WithHTTPClient(kratosHTTP))
	kratosIdentities := kratosclient.NewClient(cfg.KratosAdminURL, kratosclient.WithHTTPClient(kratosHTTP))
	keto := ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL, ketoclient.WithHTTPClient(&http.Client{
		Transport: transport,
		Timeout:   cfg.KetoWriteTimeout,
	}))

	startupContext, cancelStartup := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancelStartup()

	pool, err := postgres.Open(startupContext, cfg.Postgres.WithApplicationName(cfg.ServiceName).URL(), cfg.PostgresMaxConnections)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %s", cfg.Postgres.Redact(err))
	}

	defer func() {
		if !started {
			pool.Close()
		}
	}()

	valkeyClient, valkeyErr := valkeyinfra.Open(ctx, cfg.ValkeyURL, cfg.ValkeyTimeout)
	if valkeyClient == nil {
		return nil, valkeyErr
	}
	defer func() {
		if !started {
			valkeyClient.Close()
		}
	}()
	if valkeyErr != nil {
		logger.Warn("valkey unavailable at startup; starting in degraded mode", "error", valkeyErr)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
	}

	metrics.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		newPGXPoolCollector(pool),
	)

	ketoHTTP := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}
	ketoReader := ketoclient.NewReadClient(cfg.KetoReadURL, ketoclient.WithHTTPClient(ketoHTTP))
	permissionChecker := permissions.NewKetoChecker(ketoReader)
	authzService := featureauthz.NewService(permissionChecker)
	announcementsRepository := announcements.NewAnnouncementsRepository(pool)
	languagesRepository := languages.NewLanguagesRepository(pool)
	pagesRepository := pages.NewPagesRepository(pool)
	postsRepository := posts.NewPostsRepository(pool)
	profileRepository := profile.NewProfileRepository(pool)
	userCache := profile.NewUserCache(kratosIdentities, profileRepository)
	announcementsService := announcements.NewService(announcementsRepository)
	languagesService := languages.NewService(languagesRepository)
	pagesService := pages.NewService(pagesRepository)
	postsService := posts.NewService(postsRepository)
	profileService := profile.NewService(userCache, commonroles.NewKetoService(ketoReader, "app", "tadoku"), permissionChecker)
	api := app.New(announcementsService, authzService, languagesService, pagesService, postsService, profileService, pool, permissionChecker)
	rejectBanned := newBannedUserMiddleware(ketoReader, logger)

	handler, err := transporthttp.NewHandler(api, pool.Ping, cfg.RequestTimeout, metrics, logger, authenticate, rejectBanned)
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
		logger,
	)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
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
		ctx:             ctx,
		cancel:          cancel,
		server:          server,
		listener:        listener,
		serverErrors:    make(chan error, 2),
		metricsServer:   metricsServer,
		metricsListener: metricsListener,
		shutdownTimeout: cfg.ShutdownTimeout,
		transport:       transport,
		pool:            pool,
		valkey:          valkeyClient,
		kratos:          kratos,
		keto:            keto,
		userCacheDone:   make(chan struct{}),
	}

	go func() {
		defer close(app.userCacheDone)
		userCache.Run(ctx)
	}()

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

func newBannedUserMiddleware(keto ketoclient.AuthorizationReader, logger *slog.Logger) func(http.Handler) http.Handler {
	return transporthttp.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return keto.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
}

func (app *application) wait() error {
	var runErr error
	select {
	case <-app.ctx.Done():
	case runErr = <-app.serverErrors:
	}
	app.cancel()
	if app.userCacheDone != nil {
		<-app.userCacheDone
	}

	// Each server gets the full configured grace period.
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), app.shutdownTimeout)
	shutdownErr := app.server.Shutdown(shutdownContext)
	cancelShutdown()
	if shutdownErr != nil {
		// Cancel request contexts that outlasted graceful shutdown.
		_ = app.server.Close()
	}

	metricsContext, cancelMetrics := context.WithTimeout(context.Background(), app.shutdownTimeout)
	metricsErr := app.metricsServer.Shutdown(metricsContext)
	cancelMetrics()
	if metricsErr != nil {
		_ = app.metricsServer.Close()
	}

	app.pool.Close()
	app.valkey.Close()
	app.transport.CloseIdleConnections()

	return errors.Join(runErr, shutdownErr, metricsErr)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	app, err := start(ctx, cfg, logger)
	if err != nil {
		logger.Error("start tadoku-api", "error", err)
		os.Exit(1)
	}

	if err := app.wait(); err != nil {
		logger.Error("stop tadoku-api", "error", err)
		os.Exit(1)
	}
}
