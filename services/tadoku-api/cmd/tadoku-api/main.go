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
	fliptclient "github.com/tadoku/tadoku/services/common/client/flipt"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/client/s2s"
	"github.com/tadoku/tadoku/services/common/featureflags"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	featureaudit "github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	featureflagsservice "github.com/tadoku/tadoku/services/tadoku-api/features/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	valkeyinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/storage/postgres/asyncoutbox"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
	valkeygo "github.com/valkey-io/valkey-go"
)

type config struct {
	ScoringEngineEnabled    bool          `envconfig:"scoring_engine_enabled" required:"true"`
	Port                    int           `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort             int           `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName             string        `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	JWKS                    string        `validate:"required"`
	JWTIssuer               string        `envconfig:"jwt_issuer"`
	KetoReadURL             string        `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL            string        `validate:"required" envconfig:"keto_write_url"`
	KetoWriteTimeout        time.Duration `validate:"gt=0" envconfig:"keto_write_timeout" default:"2s"`
	OathkeeperAuthzToken    string        `validate:"required" envconfig:"oathkeeper_authz_token"`
	OathkeeperURL           string        `envconfig:"oathkeeper_url" default:"http://oathkeeper-proxy.default:4455"`
	ServiceAccountTokenPath string        `validate:"required" envconfig:"service_account_token_path" default:"/var/run/secrets/tokens/token"`

	FliptEnabled        bool          `envconfig:"flipt_enabled" default:"false"`
	FliptURL            string        `envconfig:"flipt_url" default:"http://oathkeeper-proxy.default:4455/flipt"`
	FliptEnvironment    string        `envconfig:"flipt_environment" default:"local"`
	FliptNamespace      string        `envconfig:"flipt_namespace" default:"default"`
	FliptUpdateInterval time.Duration `validate:"gte=1s" envconfig:"flipt_update_interval" default:"30s"`
	FliptRequestTimeout time.Duration `validate:"gte=1s" envconfig:"flipt_request_timeout" default:"5s"`
	FliptStartupTimeout time.Duration `validate:"gt=0" envconfig:"flipt_startup_timeout" default:"3s"`
	FliptManagementURL  string        `envconfig:"flipt_management_url" default:"http://oathkeeper-proxy.default:4455/flipt-management"`

	KratosAdminURL string        `validate:"required" envconfig:"kratos_admin_url"`
	KratosTimeout  time.Duration `validate:"gt=0" envconfig:"kratos_timeout" default:"2s"`

	PostgresMaxConnections     int32                 `validate:"gt=0,lte=32" envconfig:"postgres_max_connections" default:"4"`
	Postgres                   postgresconfig.Config `ignored:"true"`
	ValkeyURL                  string                `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout              time.Duration         `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`
	LeaderboardOutboxEnabled   bool                  `envconfig:"leaderboard_outbox_enabled" default:"false"`
	LeaderboardSharedReadiness bool                  `envconfig:"leaderboard_shared_readiness" default:"false"`
	LeaderboardCachePrefix     string                `envconfig:"leaderboard_cache_prefix"`

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
	if cfg.LeaderboardCachePrefix != "" {
		if !strings.HasSuffix(cfg.LeaderboardCachePrefix, ":") || strings.IndexFunc(cfg.LeaderboardCachePrefix, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == ':')
		}) >= 0 {
			return config{}, fmt.Errorf("validate config: LeaderboardCachePrefix must contain only lowercase letters, digits, hyphens and colons, and end in a colon")
		}
		if !cfg.LeaderboardOutboxEnabled && !cfg.LeaderboardSharedReadiness {
			return config{}, fmt.Errorf("validate config: LeaderboardCachePrefix requires a leaderboard readiness authority")
		}
	}
	if cfg.FliptEnabled {
		if strings.TrimSpace(cfg.FliptEnvironment) == "" {
			return config{}, fmt.Errorf("validate config: FliptEnvironment is required")
		}
		if strings.TrimSpace(cfg.FliptNamespace) == "" {
			return config{}, fmt.Errorf("validate config: FliptNamespace is required")
		}

		var validationErr error
		cfg.OathkeeperURL, validationErr = providerBaseURL("OathkeeperURL", cfg.OathkeeperURL)
		if validationErr != nil {
			return config{}, validationErr
		}
		cfg.FliptURL, validationErr = providerBaseURL("FliptURL", cfg.FliptURL)
		if validationErr != nil {
			return config{}, validationErr
		}
		cfg.FliptManagementURL, validationErr = providerBaseURL("FliptManagementURL", cfg.FliptManagementURL)
		if validationErr != nil {
			return config{}, validationErr
		}
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

func providerBaseURL(name, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Hostname() == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
		parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.ForceQuery || strings.Contains(rawURL, "#") {
		return "", fmt.Errorf("validate config: %s must be an HTTP(S) URL without credentials, query or fragment", name)
	}
	return strings.TrimRight(rawURL, "/"), nil
}

type application struct {
	ctx          context.Context
	cancel       context.CancelFunc
	server       *http.Server
	listener     net.Listener
	serverErrors chan error
	workerDone   chan struct{}

	metricsServer   *http.Server
	metricsListener net.Listener

	shutdownTimeout time.Duration
	transport       *http.Transport
	pool            *pgxpool.Pool
	valkey          valkeygo.Client
	kratos          *kratosapi.APIClient
	keto            *ketoclient.Client
	flipt           *fliptclient.Client
	fliptEvaluation *http.Client
	fliptManagement *http.Client
}

func noRedirect(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }

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
	authenticateCallback, err := transporthttp.NewCallbackAuthentication(cfg.OathkeeperAuthzToken)
	if err != nil {
		return nil, err
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

	exchangeHTTP := &http.Client{
		Transport:     transport,
		Timeout:       cfg.FliptRequestTimeout,
		CheckRedirect: noRedirect,
	}
	s2sClient := s2s.NewClient(
		cfg.OathkeeperURL,
		s2s.WithHTTPClient(exchangeHTTP),
		s2s.WithTokenPath(cfg.ServiceAccountTokenPath),
	)
	fliptEvaluation := &http.Client{
		Transport:     s2s.NewAuthTransport(s2sClient, "flipt-evaluation/tadoku-api", transport),
		Timeout:       cfg.FliptRequestTimeout,
		CheckRedirect: noRedirect,
	}
	fliptManagement := &http.Client{
		Transport:     s2s.NewAuthTransport(s2sClient, "flipt-management/tadoku-api", transport),
		Timeout:       cfg.FliptRequestTimeout,
		CheckRedirect: noRedirect,
	}

	featureFlagMetrics := featureflags.NewMetrics(metrics)
	var fliptProvider *fliptclient.Client
	if cfg.FliptEnabled {
		fliptProvider, err = fliptclient.New(ctx, fliptclient.Config{
			URL:            cfg.FliptURL,
			Environment:    cfg.FliptEnvironment,
			Namespace:      cfg.FliptNamespace,
			UpdateInterval: cfg.FliptUpdateInterval,
			RequestTimeout: cfg.FliptRequestTimeout,
			StartupTimeout: cfg.FliptStartupTimeout,
			HTTPClient:     fliptEvaluation,
		}, featureFlagMetrics)
		if err != nil {
			logger.Warn("feature flag provider unavailable; using safe defaults", "error", err)
			fliptProvider = nil
		}
	}
	defer func() {
		if !started && fliptProvider != nil {
			closeCtx, cancelClose := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancelClose()
			_ = fliptProvider.Close(closeCtx)
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
		postgres.NewPoolCollector(pool),
	)

	ketoHTTP := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}
	ketoReader := ketoclient.NewReadClient(cfg.KetoReadURL, ketoclient.WithHTTPClient(ketoHTTP))
	permissionChecker := permissions.NewKetoChecker(ketoReader)
	roleService := commonroles.NewKetoService(ketoReader, "app", "tadoku")
	authzService := featureauthz.NewService(
		permissionChecker,
		kratosIdentities,
		roleService,
		commonroles.NewKetoManager(keto, "app", "tadoku"),
		nil,
	)
	auditService := featureaudit.NewService(featureaudit.NewRepository(pool))
	announcementsRepository := announcements.NewAnnouncementsRepository(pool)
	contestsRepository := contests.NewContestsRepository(pool)
	outboxRepository := asyncoutbox.NewRepository(pool)
	languagesRepository := languages.NewLanguagesRepository(pool)
	leaderboardRepository := leaderboard.NewRepository(pool)
	logsRepository := logs.NewLogsRepository(pool)
	pagesRepository := pages.NewPagesRepository(pool)
	postsRepository := posts.NewPostsRepository(pool)
	profileRepository := profile.NewRepository(pool)
	scoringRepository := scoring.NewScoringRepository(pool)
	userCache := profile.NewUserCache(kratosIdentities)
	announcementsService := announcements.NewService(announcementsRepository)
	contestsService := contests.NewService(contestsRepository, outboxRepository)
	languagesService := languages.NewService(languagesRepository)
	leaderboardService := leaderboard.NewService(leaderboardRepository, valkeyClient, cfg.ValkeyTimeout, cfg.LeaderboardCachePrefix)
	if cfg.LeaderboardSharedReadiness {
		leaderboardService.EnableSharedReadiness()
	}
	logsService := logs.NewService(logsRepository, outboxRepository, cfg.ScoringEngineEnabled)
	pagesService := pages.NewService(pagesRepository)
	postsService := posts.NewService(postsRepository)
	profileService := profile.NewService(profileRepository, userCache, roleService, kratosIdentities)
	featureFlagService := featureflagsservice.NewService(featureflags.NewEvaluator(fliptProvider, featureFlagMetrics), fliptmanagement.NewClient(fliptmanagement.Config{
		URL:         cfg.FliptManagementURL,
		Environment: cfg.FliptEnvironment,
		HTTPClient:  fliptManagement,
	}))
	scoringObserver := observability.NewScoringObserver(metrics, logger, cfg.ScoringEngineEnabled)
	scoringService := scoring.NewService(scoringRepository, cfg.ScoringEngineEnabled, scoringObserver)
	api := app.New(app.Dependencies{
		Announcements: announcementsService,
		Audit:         auditService,
		Authorization: authzService,
		Contests:      contestsService,
		Leaderboard:   leaderboardService,
		Languages:     languagesService,
		Logs:          logsService,
		Pages:         pagesService,
		Posts:         postsService,
		Profile:       profileService,
		Scoring:       scoringService,
		FeatureFlags:  featureFlagService,
		DB:            pool,
		Permissions:   permissionChecker,
	})
	rejectBanned := transporthttp.RejectBannedUsers(permissionChecker.CheckBanned, logger)

	handler, err := transporthttp.NewHandler(api, pool.Ping, cfg.RequestTimeout, metrics, logger, authenticate, rejectBanned, authenticateCallback)
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
		flipt:           fliptProvider,
		fliptEvaluation: fliptEvaluation,
		fliptManagement: fliptManagement,
	}
	if cfg.LeaderboardOutboxEnabled {
		app.workerDone = make(chan struct{})
		worker := leaderboard.NewWorker(leaderboardService, logger)
		go func() {
			defer close(app.workerDone)
			worker.Run(ctx)
		}()
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

func (app *application) wait() error {
	var runErr error
	select {
	case <-app.ctx.Done():
	case runErr = <-app.serverErrors:
	}
	app.cancel()
	if app.workerDone != nil {
		<-app.workerDone
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), app.shutdownTimeout)
	shutdownErr := app.server.Shutdown(shutdownContext)
	cancelShutdown()
	if shutdownErr != nil {
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
	var fliptErr error
	if app.flipt != nil {
		closeContext, cancelClose := context.WithTimeout(context.Background(), app.shutdownTimeout)
		fliptErr = app.flipt.Close(closeContext)
		cancelClose()
	}
	app.transport.CloseIdleConnections()

	return errors.Join(runErr, shutdownErr, metricsErr, fliptErr)
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
