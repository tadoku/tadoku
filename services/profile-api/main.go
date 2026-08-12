package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/common/health"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	commonobservability "github.com/tadoku/tadoku/services/common/observability"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/profile-api/cache"
	"github.com/tadoku/tadoku/services/profile-api/client/ory"
	profiledomain "github.com/tadoku/tadoku/services/profile-api/domain"
	"github.com/tadoku/tadoku/services/profile-api/http/rest"
	"github.com/tadoku/tadoku/services/profile-api/http/rest/openapi"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port                   int64   `validate:"required"`
	JWKS                   string  `validate:"required"`
	KratosURL              string  `validate:"required" envconfig:"kratos_url"`
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	ServiceName            string  `envconfig:"service_name" default:"profile-api"`
	MetricsPort            int64   `envconfig:"metrics_port" default:"9090"`
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
	userCache := cache.NewUserCache(kratosClient, 5*time.Minute)
	userCache.Start()

	rolesSvc := commonroles.NewKetoService(ketoclient.NewReadClient(cfg.KetoReadURL), "app", "tadoku")
	serviceMetrics := commonobservability.NewMetrics(psql, cfg.ServiceName)
	metricsServer := commonobservability.NewServer(
		fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort),
		serviceMetrics.Handler(),
	)
	if err := metricsServer.Start(); err != nil {
		panic(fmt.Errorf("could not start internal metrics server: %w", err))
	}

	userList := profiledomain.NewUserList(userCache, rolesSvc)

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

	server := rest.NewServer(userList)
	internalServer := rest.NewInternalServer()

	openapi.RegisterHandlersWithBaseURL(api, server, "")
	internal := api.Group("", tadokumiddleware.RequireServiceIdentity())
	rest.RegisterInternalRoutes(internal, internalServer)

	defer userCache.Stop()

	go func() {
		fmt.Printf("profile-api is now available at: http://localhost:%d/v2\n", cfg.Port)
		if err := e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case metricsErr := <-metricsServer.Errors():
		slog.Error("internal metrics server stopped", "error", metricsErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("could not gracefully shut down application server", "error", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		slog.Error("could not gracefully shut down metrics server", "error", err)
	}
}
