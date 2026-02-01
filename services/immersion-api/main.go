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
	"github.com/tadoku/tadoku/services/common/domain"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/common/storage/memory"
	"github.com/tadoku/tadoku/services/immersion-api/cache"
	"github.com/tadoku/tadoku/services/immersion-api/client/ory"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
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
	roleRepository := memory.NewRoleRepository("/etc/tadoku/permissions/roles.yaml")

	e := echo.New()
	e.Use(tadokumiddleware.Logger([]string{"/ping"}))
	e.Use(tadokumiddleware.SessionJWT(cfg.JWKS))
	e.Use(tadokumiddleware.Session(roleRepository))
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

	commandService := command.NewService(postgresRepository, clock)
	queryService := query.NewService(postgresRepository, clock, kratosClient, userCache)

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

	server := rest.NewServer(
		commandService,
		queryService,
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
