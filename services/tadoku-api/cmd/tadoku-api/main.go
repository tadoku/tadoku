package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tadoku/tadoku/services/common/client/s2s"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	featureaudit "github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	featureflagsservice "github.com/tadoku/tadoku/services/tadoku-api/features/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	valkeyinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
	valkeygo "github.com/valkey-io/valkey-go"
)

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
	keto            *ketoclient.Client
	flipt           *fliptclient.Provider
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
	authenticate, err := transporthttp.NewJWTAuthentication(ctx, cfg.JWKS, cfg.DialTimeout, cfg.MaxTokenAge, cfg.JWTIssuer, cfg.deployment, logger)
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
	targets, err := fliptclient.NewTargets(cfg.FliptEnvironment, cfg.FliptNamespace, cfg.FliptTestEnvironment)
	if err != nil && cfg.FliptEnabled {
		return nil, err
	}
	var fliptProvider *fliptclient.Provider
	if cfg.FliptEnabled {
		fliptProvider, err = fliptclient.NewProvider(ctx, fliptclient.Config{
			URL:            cfg.FliptURL,
			Environment:    cfg.FliptEnvironment,
			Namespace:      cfg.FliptNamespace,
			UpdateInterval: cfg.FliptUpdateInterval,
			RequestTimeout: cfg.FliptRequestTimeout,
			StartupTimeout: cfg.FliptStartupTimeout,
			HTTPClient:     fliptEvaluation,
		}, targets, featureFlagMetrics)
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

	keto := ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL, ketoclient.WithHTTPClient(&http.Client{
		Transport: transport,
		Timeout:   cfg.KetoWriteTimeout,
	}))

	startupContext, cancelStartup := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancelStartup()

	pool, err := postgres.Open(
		startupContext,
		cfg.Postgres.WithApplicationName(cfg.ServiceName).URL().Reveal(),
		cfg.PostgresMaxConnections,
	)
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
	roleService := permissions.NewKetoService(ketoReader)

	identityReader := identity.NewReader(pool)
	authzService := featureauthz.NewService(
		permissionChecker,
		identityReader,
		roleService,
		permissions.NewKetoManager(keto),
		nil,
	)

	auditService := featureaudit.NewService(featureaudit.NewRepository(pool))
	announcementsRepository := announcements.NewAnnouncementsRepository(pool)
	contestsRepository := contests.NewContestsRepository(pool)
	jobQueue := jobqueue.NewService(jobqueue.NewRepository(pool))
	languagesRepository := languages.NewLanguagesRepository(pool)
	leaderboardRepository := leaderboard.NewRepository(pool)
	logsRepository := logs.NewLogsRepository(pool)
	pagesRepository := pages.NewPagesRepository(pool)
	postsRepository := posts.NewPostsRepository(pool)
	profileRepository := profile.NewRepository(pool)
	scoringRepository := scoring.NewScoringRepository(pool)

	announcementsService := announcements.NewService(announcementsRepository)
	contestsService := contests.NewService(contestsRepository)
	languagesService := languages.NewService(languagesRepository)
	leaderboardService := leaderboard.NewService(
		leaderboardRepository,
		leaderboard.NewCache(valkeyClient, cfg.ValkeyTimeout),
	)
	logsService := logs.NewService(logsRepository, cfg.ScoringEngineEnabled)
	pagesService := pages.NewService(pagesRepository)
	postsService := posts.NewService(postsRepository)
	profileService := profile.NewService(profileRepository, roleService, identityReader)
	featureFlagService := featureflagsservice.NewService(
		featureflags.NewEvaluator(fliptProvider, featureFlagMetrics),
		fliptmanagement.NewClient(fliptmanagement.Config{
			URL:        cfg.FliptManagementURL,
			Targets:    targets,
			HTTPClient: fliptManagement,
		}),
	)

	scoringObserver := observability.NewScoringObserver(metrics, logger, cfg.ScoringEngineEnabled)
	scoringService := scoring.NewService(scoringRepository, cfg.ScoringEngineEnabled, scoringObserver)
	identityService := identity.NewService(
		identity.NewRepository(pool),
		identity.NewThrottle(valkeyClient, cfg.ValkeyTimeout, logger),
		cfg.IdentityCookieDomain,
	)

	api := app.New(app.Dependencies{
		Tokens:        tokens.NewService(tokens.NewRepository(pool), cfg.Production),
		Deployment:    cfg.deployment,
		Email:         email.NewService(email.NewRepository(pool), nil),
		JobQueue:      jobQueue,
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
		Identity:      identityService,
	})
	rejectBanned := transporthttp.RejectBannedUsers(permissionChecker.Admit, logger)

	handler, err := transporthttp.NewHandler(
		api,
		pool.Ping,
		cfg.RequestTimeout,
		metrics,
		logger,
		authenticate,
		rejectBanned,
		authenticateCallback,
		transporthttp.WithIdentityTrustedProxies(cfg.identityTrustedProxies),
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
		keto:            keto,
		flipt:           fliptProvider,
		fliptEvaluation: fliptEvaluation,
		fliptManagement: fliptManagement,
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
	logger := slog.New(observability.NewTenantHandler(slog.NewJSONHandler(os.Stdout, nil)))
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
