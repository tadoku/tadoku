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
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/health"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	commonobservability "github.com/tadoku/tadoku/services/common/observability"
	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/content-api/storage/postgres"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port                   int64   `validate:"required"`
	JWKS                   string  `validate:"required"`
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	ServiceName            string  `envconfig:"service_name" default:"content-api"`
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

	pageRepository := postgres.NewPageRepository(psql)
	postRepository := postgres.NewPostRepository(psql)
	announcementRepository := postgres.NewAnnouncementRepository(psql)
	rolesSvc := commonroles.NewKetoService(ketoclient.NewReadClient(cfg.KetoReadURL), "app", "tadoku")
	serviceMetrics := commonobservability.NewMetrics(psql, cfg.ServiceName)
	metricsServer := commonobservability.NewServer(
		fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort),
		serviceMetrics.Handler(),
	)
	if err := metricsServer.Start(); err != nil {
		panic(fmt.Errorf("could not start internal metrics server: %w", err))
	}

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

	clock, err := commondomain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	// Page services
	pageCreate := domain.NewPageCreate(pageRepository, clock)
	pageUpdate := domain.NewPageUpdate(pageRepository, clock)
	pageDelete := domain.NewPageDelete(pageRepository)
	pageFind := domain.NewPageFind(pageRepository, clock)
	pageFindByID := domain.NewPageFindByID(pageRepository)
	pageList := domain.NewPageList(pageRepository)
	pageVersionList := domain.NewPageVersionList(pageRepository)
	pageVersionGet := domain.NewPageVersionGet(pageRepository)

	// Post services
	postCreate := domain.NewPostCreate(postRepository, clock)
	postUpdate := domain.NewPostUpdate(postRepository, clock)
	postDelete := domain.NewPostDelete(postRepository)
	postFind := domain.NewPostFind(postRepository, clock)
	postFindByID := domain.NewPostFindByID(postRepository)
	postList := domain.NewPostList(postRepository)
	postVersionList := domain.NewPostVersionList(postRepository)
	postVersionGet := domain.NewPostVersionGet(postRepository)

	// Announcement services
	announcementCreate := domain.NewAnnouncementCreate(announcementRepository, clock)
	announcementUpdate := domain.NewAnnouncementUpdate(announcementRepository, clock)
	announcementDelete := domain.NewAnnouncementDelete(announcementRepository)
	announcementFindByID := domain.NewAnnouncementFindByID(announcementRepository)
	announcementList := domain.NewAnnouncementList(announcementRepository)
	announcementListActive := domain.NewAnnouncementListActive(announcementRepository)

	server := rest.NewServer(
		pageCreate,
		pageUpdate,
		pageDelete,
		pageFind,
		pageFindByID,
		pageList,
		pageVersionList,
		pageVersionGet,
		postCreate,
		postUpdate,
		postDelete,
		postFind,
		postFindByID,
		postList,
		postVersionList,
		postVersionGet,
		announcementCreate,
		announcementUpdate,
		announcementDelete,
		announcementFindByID,
		announcementList,
		announcementListActive,
	)

	openapi.RegisterHandlersWithBaseURL(api, server, "")

	go func() {
		fmt.Printf("content-api is now available at: http://localhost:%d/v2\n", cfg.Port)
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
