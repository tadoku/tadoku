package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	fliptclient "github.com/tadoku/tadoku/services/common/client/flipt"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/client/s2s"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/featureflags"
	"github.com/tadoku/tadoku/services/common/health"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	commonobservability "github.com/tadoku/tadoku/services/common/observability"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/immersion-api/client/ory"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/observability"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/repository"
	valkeystore "github.com/tadoku/tadoku/services/immersion-api/storage/valkey"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valkey-io/valkey-go"
)

type Config struct {
	Port                   int64         `validate:"required"`
	JWKS                   string        `validate:"required"`
	KratosURL              string        `validate:"required" envconfig:"kratos_url"`
	OathkeeperURL          string        `validate:"required" envconfig:"oathkeeper_url"`
	KetoReadURL            string        `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL           string        `validate:"required" envconfig:"keto_write_url"`
	ValkeyURL              string        `validate:"required" envconfig:"valkey_url"`
	ValkeyTimeout          time.Duration `validate:"gt=0" envconfig:"valkey_timeout" default:"1s"`
	ServiceName            string        `envconfig:"service_name" default:"immersion-api"`
	SentryDSN              string        `envconfig:"sentry_dns"`
	SentryTracesSampleRate float64       `validate:"required_with=SentryDSN" envconfig:"sentry_traces_sample_rate"`
	ScoringEngineEnabled   bool          `envconfig:"scoring_engine_enabled" default:"false"`
	MetricsPort            int64         `envconfig:"metrics_port" default:"9090"`
	FliptEnabled           bool          `envconfig:"flipt_enabled" default:"false"`
	FliptURL               string        `envconfig:"flipt_url" default:"http://oathkeeper-proxy.default:4455/flipt"`
	FliptEnvironment       string        `envconfig:"flipt_environment" default:"local"`
	FliptNamespace         string        `envconfig:"flipt_namespace" default:"default"`
	FliptUpdateInterval    time.Duration `envconfig:"flipt_update_interval" default:"30s"`
	FliptRequestTimeout    time.Duration `envconfig:"flipt_request_timeout" default:"5s"`
	FliptStartupTimeout    time.Duration `envconfig:"flipt_startup_timeout" default:"3s"`
}

type featureFlagProviderInitializer func() (*fliptclient.Client, error)

type valkeyClientInitializer func(valkey.ClientOption) (valkey.Client, error)

func initializeValkeyClient(cfg Config, initialize valkeyClientInitializer) (valkey.Client, error) {
	if cfg.ValkeyTimeout <= 0 {
		return nil, fmt.Errorf("valkey timeout must be positive")
	}

	option, err := valkey.ParseURL(cfg.ValkeyURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse valkey url: %w", err)
	}
	if len(option.InitAddress) != 1 {
		return nil, fmt.Errorf("valkey url must configure exactly one standalone address")
	}
	option.Dialer.Timeout = cfg.ValkeyTimeout
	option.ForceSingleClient = true
	option.DisableRetry = true

	client, err := initialize(option)
	if client == nil {
		if err == nil {
			err = fmt.Errorf("initializer returned a nil client")
		}
		return nil, fmt.Errorf("could not connect to valkey: %w", err)
	}
	if err != nil {
		slog.Warn("valkey unavailable at startup; starting in degraded mode", "error", err)
	}
	return client, nil
}

func initializeFeatureFlagProvider(cfg Config, initialize featureFlagProviderInitializer) (*fliptclient.Client, error) {
	if !cfg.FliptEnabled {
		return nil, nil
	}
	return initialize()
}

func main() {
	cfg := Config{}
	if err := envconfig.Process("API", &cfg); err != nil {
		panic(fmt.Errorf("could not configure server: %w", err))
	}

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		panic(fmt.Errorf("could not configure server: %w", err))
	}

	postgresConfig, err := postgresconfig.Load("API_POSTGRES", "API_POSTGRES_URL")
	if err != nil {
		panic(fmt.Errorf("could not configure postgres: %w", err))
	}
	connConfig, err := postgresConfig.ConnConfig()
	if err != nil {
		panic(err)
	}
	psql := stdlib.OpenDB(*connConfig)

	kratosClient := ory.NewKratosClient(cfg.KratosURL)

	postgresRepository := repository.NewRepository(psql)
	var ketoAuthz ketoclient.AuthorizationClient = ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL)
	rolesSvc := commonroles.NewKetoService(ketoAuthz, "app", "tadoku")

	valkeyClient, err := initializeValkeyClient(cfg, valkey.NewClient)
	if err != nil {
		panic(err)
	}
	defer valkeyClient.Close()

	clock, err := domain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	serviceMetrics := commonobservability.NewMetrics(psql, cfg.ServiceName)
	serviceLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	leaderboardStore := valkeystore.NewLeaderboardStore(
		valkeyClient,
		clock,
		cfg.ValkeyTimeout,
	)
	leaderboardUpdater := immersiondomain.NewLeaderboardUpdater(leaderboardStore, postgresRepository)
	featureFlagMetrics := featureflags.NewMetrics(serviceMetrics.Registry(), clock)
	fliptProvider, err := initializeFeatureFlagProvider(cfg, func() (*fliptclient.Client, error) {
		s2sClient := s2s.NewClient(cfg.OathkeeperURL, clock)
		fliptHTTPClient := &http.Client{
			Transport: s2s.NewAuthTransport(s2sClient, "flipt-evaluation/immersion-api", http.DefaultTransport),
		}
		return fliptclient.New(context.Background(), fliptclient.Config{
			URL:            cfg.FliptURL,
			Environment:    cfg.FliptEnvironment,
			Namespace:      cfg.FliptNamespace,
			UpdateInterval: cfg.FliptUpdateInterval,
			RequestTimeout: cfg.FliptRequestTimeout,
			StartupTimeout: cfg.FliptStartupTimeout,
			HTTPClient:     fliptHTTPClient,
		}, featureFlagMetrics)
	})
	if err != nil {
		slog.Warn("feature flag provider unavailable; using safe defaults")
	}
	if fliptProvider != nil {
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := fliptProvider.Close(closeCtx); err != nil {
				slog.Warn("feature flag provider shutdown failed")
			}
		}()
	}
	featureFlagEvaluator := featureflags.NewEvaluator(fliptProvider, featureFlagMetrics, clock)
	scoringMetrics := observability.NewScoringShadowMetrics(serviceMetrics.Registry(), cfg.ScoringEngineEnabled)
	scoringObserver := observability.NewScoringShadowObserver(scoringMetrics, serviceLogger)
	metricsServer := commonobservability.NewServer(
		fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort),
		serviceMetrics.Handler(),
	)
	if err := metricsServer.Start(); err != nil {
		panic(fmt.Errorf("could not start internal metrics server: %w", err))
	}

	// Start leaderboard outbox worker for async leaderboard sync
	outboxWorker := immersiondomain.NewLeaderboardOutboxWorker(postgresRepository, leaderboardUpdater, clock, 500*time.Millisecond)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go outboxWorker.Run(workerCtx)

	e := echo.New()
	e.Use(serviceMetrics.Middleware())
	e.Use(middleware.Recover())

	// Health endpoints: allow K8s probes without auth, require admin if JWT is present
	optAuth := tadokumiddleware.OptionalAdminAuth(cfg.JWKS, rolesSvc)
	pgChecker := health.NewPostgresChecker("postgres", psql)
	e.GET("/livez", health.LivezHandler, optAuth)
	e.GET("/readyz", health.ReadyzHandler([]health.HealthChecker{pgChecker}), optAuth)

	// Business endpoints: full auth middleware stack
	api := e.Group("")
	api.Use(tadokumiddleware.Logger([]string{"/ping"}))
	api.Use(tadokumiddleware.VerifyJWT(cfg.JWKS))
	api.Use(tadokumiddleware.Identity())
	api.Use(tadokumiddleware.RolesFromKeto(rolesSvc))
	api.Use(tadokumiddleware.RequireServiceAudience(cfg.ServiceName))
	api.Use(tadokumiddleware.RejectBannedUsers())

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: cfg.SentryTracesSampleRate,
		}); err != nil {
			panic(fmt.Errorf("sentry initialization failed: %v", err))
		}
		api.Use(sentryecho.New(sentryecho.Options{}))
	}

	// Service-per-function services
	contestConfigurationOptions := immersiondomain.NewContestConfigurationOptions(postgresRepository)
	logConfigurationOptions := immersiondomain.NewLogConfigurationOptionsWithScoringEngine(postgresRepository, cfg.ScoringEngineEnabled)
	contestFindLatestOfficial := immersiondomain.NewContestFindLatestOfficial(postgresRepository)
	contestSummaryFetch := immersiondomain.NewContestSummaryFetch(postgresRepository)
	profileYearlyActivitySplit := immersiondomain.NewProfileYearlyActivitySplit(postgresRepository)
	contestFind := immersiondomain.NewContestFind(postgresRepository)
	logFind := immersiondomain.NewLogFind(postgresRepository)
	contestList := immersiondomain.NewContestList(postgresRepository)
	logListForUser := immersiondomain.NewLogListForUser(postgresRepository)
	logListForContest := immersiondomain.NewLogListForContest(postgresRepository)
	registrationFind := immersiondomain.NewRegistrationFind(postgresRepository)
	registrationListYearly := immersiondomain.NewRegistrationListYearly(postgresRepository)
	contestLeaderboardFetch := immersiondomain.NewContestLeaderboardFetch(postgresRepository, leaderboardStore)
	leaderboardYearly := immersiondomain.NewLeaderboardYearly(postgresRepository, leaderboardStore)
	leaderboardGlobal := immersiondomain.NewLeaderboardGlobal(postgresRepository, leaderboardStore)
	profileContest := immersiondomain.NewProfileContest(postgresRepository)
	profileContestActivity := immersiondomain.NewProfileContestActivity(postgresRepository)
	profileYearlyActivity := immersiondomain.NewProfileYearlyActivity(postgresRepository)
	profileYearlyScores := immersiondomain.NewProfileYearlyScores(postgresRepository)
	profileFetch := immersiondomain.NewProfileFetch(kratosClient)
	registrationListOngoing := immersiondomain.NewRegistrationListOngoing(postgresRepository, clock)
	contestPermissionCheck := immersiondomain.NewContestPermissionCheck(postgresRepository, kratosClient, clock)
	logDelete := immersiondomain.NewLogDelete(postgresRepository, clock)
	contestModerationDetachLog := immersiondomain.NewContestModerationDetachLog(postgresRepository)
	userUpsert := immersiondomain.NewUserUpsert(postgresRepository)
	registrationUpsert := immersiondomain.NewRegistrationUpsert(postgresRepository, userUpsert)
	logCreate := immersiondomain.NewLogCreateWithScoringObserver(postgresRepository, clock, userUpsert, cfg.ScoringEngineEnabled, scoringObserver)
	logUpdate := immersiondomain.NewLogUpdateWithScoringObserver(postgresRepository, clock, cfg.ScoringEngineEnabled, scoringObserver)
	contestCreate := immersiondomain.NewContestCreate(postgresRepository, clock, userUpsert)
	languageList := immersiondomain.NewLanguageList(postgresRepository)
	languageCreate := immersiondomain.NewLanguageCreate(postgresRepository)
	languageUpdate := immersiondomain.NewLanguageUpdate(postgresRepository)
	tagSuggestions := immersiondomain.NewTagSuggestions(postgresRepository)
	logContestUpdate := immersiondomain.NewLogContestUpdateWithScoringEngine(postgresRepository, clock, cfg.ScoringEngineEnabled)
	scorePreview := immersiondomain.NewScorePreview(postgresRepository, clock)
	scoringRuleSetManagement := immersiondomain.NewScoringRuleSetManagement(postgresRepository, clock)
	accountDeletionLock := immersiondomain.NewAccountDeletionLock(postgresRepository, clock)
	accountDeletionScrub := immersiondomain.NewAccountDeletionScrub(postgresRepository, clock)

	server := rest.NewServer(
		contestConfigurationOptions,
		logConfigurationOptions,
		contestFindLatestOfficial,
		contestSummaryFetch,
		profileYearlyActivitySplit,
		contestFind,
		logFind,
		contestList,
		logListForUser,
		logListForContest,
		registrationFind,
		registrationListYearly,
		contestLeaderboardFetch,
		leaderboardYearly,
		leaderboardGlobal,
		profileContest,
		profileContestActivity,
		profileYearlyActivity,
		profileYearlyScores,
		profileFetch,
		registrationListOngoing,
		contestPermissionCheck,
		logDelete,
		contestModerationDetachLog,
		registrationUpsert,
		logCreate,
		logUpdate,
		contestCreate,
		languageList,
		languageCreate,
		languageUpdate,
		tagSuggestions,
		logContestUpdate,
		scorePreview,
		scoringRuleSetManagement,
		featureFlagEvaluator,
	)

	openapi.RegisterHandlersWithBaseURL(api, server, "")
	internalServer := rest.NewInternalServer(accountDeletionLock, accountDeletionScrub)
	internal := api.Group("", tadokumiddleware.RequireServiceIdentity())
	rest.RegisterInternalRoutes(internal, internalServer)

	// Start server in goroutine
	go func() {
		fmt.Printf("immersion-api is now available at: http://localhost:%d/v2\n", cfg.Port)
		if err := e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case metricsErr := <-metricsServer.Errors():
		slog.Error("internal metrics server stopped", "error", metricsErr)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("could not gracefully shut down application server", "error", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		slog.Error("could not gracefully shut down metrics server", "error", err)
	}
}
