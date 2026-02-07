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
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/immersion-api/cache"
	"github.com/tadoku/tadoku/services/immersion-api/client/ory"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/repository"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	PostgresURL            string  `validate:"required" envconfig:"postgres_url"`
	Port                   int64   `validate:"required"`
	JWKS                   string  `validate:"required"`
	KratosURL              string  `validate:"required" envconfig:"kratos_url"`
	OathkeeperURL          string  `validate:"required" envconfig:"oathkeeper_url"`
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL           string  `validate:"required" envconfig:"keto_write_url"`
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
	userCache := cache.NewUserCache(kratosClient, 5*time.Minute)
	userCache.Start()

	postgresRepository := repository.NewRepository(psql)
	var ketoAuthz ketoclient.AuthorizationClient = ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL)
	rolesSvc := commonroles.NewKetoService(ketoAuthz, "app", "tadoku")
	roleMgmt := commonroles.NewKetoManager(ketoAuthz, "app", "tadoku")

	e := echo.New()
	e.Use(tadokumiddleware.Logger([]string{"/ping"}))
	e.Use(tadokumiddleware.VerifyJWT(cfg.JWKS))
	e.Use(tadokumiddleware.Identity())
	e.Use(tadokumiddleware.RolesFromKeto(rolesSvc))
	e.Use(tadokumiddleware.RequireServiceAudience(cfg.ServiceName))
	e.Use(tadokumiddleware.RejectBannedUsers())
	e.Use(middleware.Recover())

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: cfg.SentryTracesSampleRate,
		}); err != nil {
			panic(fmt.Errorf("sentry initialization failed: %v", err))
		}
		e.Use(sentryecho.New(sentryecho.Options{}))
	}

	clock, err := domain.NewClock("UTC")
	if err != nil {
		panic(err)
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
	contestLeaderboardFetch := immersiondomain.NewContestLeaderboardFetch(postgresRepository)
	leaderboardYearly := immersiondomain.NewLeaderboardYearly(postgresRepository)
	leaderboardGlobal := immersiondomain.NewLeaderboardGlobal(postgresRepository)
	profileContest := immersiondomain.NewProfileContest(postgresRepository)
	profileContestActivity := immersiondomain.NewProfileContestActivity(postgresRepository)
	profileYearlyActivity := immersiondomain.NewProfileYearlyActivity(postgresRepository)
	profileYearlyScores := immersiondomain.NewProfileYearlyScores(postgresRepository)
	profileFetch := immersiondomain.NewProfileFetch(kratosClient)
	registrationListOngoing := immersiondomain.NewRegistrationListOngoing(postgresRepository, clock)
	contestPermissionCheck := immersiondomain.NewContestPermissionCheck(postgresRepository, kratosClient, clock)
	userList := immersiondomain.NewUserList(userCache, rolesSvc)
	logDelete := immersiondomain.NewLogDelete(postgresRepository, clock)
	contestModerationDetachLog := immersiondomain.NewContestModerationDetachLog(postgresRepository)
	userUpsert := immersiondomain.NewUserUpsert(postgresRepository)
	registrationUpsert := immersiondomain.NewRegistrationUpsert(postgresRepository, userUpsert)
	logCreate := immersiondomain.NewLogCreate(postgresRepository, clock, userUpsert)
	contestCreate := immersiondomain.NewContestCreate(postgresRepository, clock, userUpsert)
	updateUserRole := immersiondomain.NewUpdateUserRole(postgresRepository, postgresRepository, rolesSvc, roleMgmt)
	languageList := immersiondomain.NewLanguageList(postgresRepository)
	languageCreate := immersiondomain.NewLanguageCreate(postgresRepository)
	languageUpdate := immersiondomain.NewLanguageUpdate(postgresRepository)

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
		userList,
		logDelete,
		contestModerationDetachLog,
		registrationUpsert,
		logCreate,
		contestCreate,
		updateUserRole,
		languageList,
		languageCreate,
		languageUpdate,
	)

	openapi.RegisterHandlersWithBaseURL(e, server, "")

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
	userCache.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
