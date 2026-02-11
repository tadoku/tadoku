package main

import (
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/profile-api/http/rest"
	"github.com/tadoku/tadoku/services/profile-api/http/rest/openapi"

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
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	ServiceName            string  `envconfig:"service_name" default:"profile-api"`
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
	_ = psql // Will be used when repositories are added

	rolesSvc := commonroles.NewKetoService(ketoclient.NewReadClient(cfg.KetoReadURL), "app", "tadoku")

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

	server := rest.NewServer()
	internalServer := rest.NewInternalServer()

	openapi.RegisterHandlersWithBaseURL(e, server, "")
	internal := e.Group("", tadokumiddleware.RequireServiceIdentity())
	rest.RegisterInternalRoutes(internal, internalServer)

	fmt.Printf("profile-api is now available at: http://localhost:%d/v2\n", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
