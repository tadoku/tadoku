package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/health"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/immersion-api/client/ory"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/repository"
	valkeystore "github.com/tadoku/tadoku/services/immersion-api/storage/valkey"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valkey-io/valkey-go"
)

type Config struct {
	PostgresURL            string  `validate:"required" envconfig:"postgres_url"`
	Port                   int64   `validate:"required"`
	JWKS                   string  `validate:"required"`
	KratosURL              string  `validate:"required" envconfig:"kratos_url"`
	OathkeeperURL          string  `validate:"required" envconfig:"oathkeeper_url"`
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL           string  `validate:"required" envconfig:"keto_write_url"`
	ValkeyURL              string  `validate:"required" envconfig:"valkey_url"`
	ServiceName            string  `envconfig:"service_name" default:"immersion-api"`
	SentryDSN              string  `envconfig:"sentry_dns"`
	SentryTracesSampleRate float64 `validate:"required_with=SentryDSN" envconfig:"sentry_traces_sample_rate"`
}

func main() {
	cfg := Config{}
	envconfig.Process("API", &cfg)

	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		panic(fmt.Errorf("could not configure server: %w", err))
	}

	psql, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		panic(err)
	}

	kratosClient := ory.NewKratosClient(cfg.KratosURL)

	postgresRepository := repository.NewRepository(psql)
	var ketoAuthz ketoclient.AuthorizationClient = ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL)
	rolesSvc := commonroles.NewKetoService(ketoAuthz, "app", "tadoku")

	valkeyOpt, err := valkey.ParseURL(cfg.ValkeyURL)
	if err != nil {
		panic(fmt.Errorf("could not parse valkey url: %w", err))
	}
	valkeyClient, err := valkey.NewClient(valkeyOpt)
	if err != nil {
		panic(fmt.Errorf("could not connect to valkey: %w", err))
	}
	defer valkeyClient.Close()

	clock, err := domain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	leaderboardStore := valkeystore.NewLeaderboardStore(valkeyClient, clock)
	leaderboardUpdater := immersiondomain.NewLeaderboardUpdater(leaderboardStore, postgresRepository)

	// Start leaderboard outbox worker for async leaderboard sync
	outboxWorker := immersiondomain.NewLeaderboardOutboxWorker(postgresRepository, leaderboardUpdater, clock, 500*time.Millisecond)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go outboxWorker.Run(workerCtx)

	e := echo.New()
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
	logConfigurationOptions := immersiondomain.NewLogConfigurationOptions(postgresRepository)
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
	logCreate := immersiondomain.NewLogCreate(postgresRepository, clock, userUpsert)
	logUpdate := immersiondomain.NewLogUpdate(postgresRepository, clock)
	contestCreate := immersiondomain.NewContestCreate(postgresRepository, clock, userUpsert)
	languageList := immersiondomain.NewLanguageList(postgresRepository)
	languageCreate := immersiondomain.NewLanguageCreate(postgresRepository)
	languageUpdate := immersiondomain.NewLanguageUpdate(postgresRepository)
	tagSuggestions := immersiondomain.NewTagSuggestions(postgresRepository)
	logContestUpdate := immersiondomain.NewLogContestUpdate(postgresRepository, clock)

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
	)

	openapi.RegisterHandlersWithBaseURL(api, server, "")

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
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
