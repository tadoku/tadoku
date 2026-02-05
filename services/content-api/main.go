package main

import (
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/common/storage/memory"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/content-api/storage/postgres"

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
	ServiceName            string  `envconfig:"service_name" default:"content-api"`
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

	pageRepository := postgres.NewPageRepository(psql)
	postRepository := postgres.NewPostRepository(psql)
	configRoleRepository := memory.NewRoleRepository("/etc/tadoku/permissions/roles.yaml")

	e := echo.New()
	e.Use(tadokumiddleware.Logger([]string{"/ping"}))
	e.Use(tadokumiddleware.VerifyJWT(cfg.JWKS))
	e.Use(tadokumiddleware.Identity(configRoleRepository, nil))
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

	clock, err := commondomain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	// Page services
	pageCreate := domain.NewPageCreate(pageRepository, clock)
	pageUpdate := domain.NewPageUpdate(pageRepository, clock)
	pageFind := domain.NewPageFind(pageRepository, clock)
	pageList := domain.NewPageList(pageRepository)

	// Post services
	postCreate := domain.NewPostCreate(postRepository, clock)
	postUpdate := domain.NewPostUpdate(postRepository, clock)
	postFind := domain.NewPostFind(postRepository, clock)
	postList := domain.NewPostList(postRepository)

	server := rest.NewServer(
		pageCreate,
		pageUpdate,
		pageFind,
		pageList,
		postCreate,
		postUpdate,
		postFind,
		postList,
	)

	openapi.RegisterHandlersWithBaseURL(e, server, "")

	fmt.Printf("content-api is now available at: http://localhost:%d/v2\n", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
