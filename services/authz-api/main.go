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
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/internalapi"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/common/health"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tadoku/tadoku/services/authz-api/storage/postgres/repository"
)

type Config struct {
	PostgresURL            string  `validate:"required" envconfig:"postgres_url"`
	Port                   int64   `validate:"required"`
	JWKS                   string  `validate:"required"`
	KratosURL              string  `validate:"required" envconfig:"kratos_url"`
	KetoReadURL            string  `validate:"required" envconfig:"keto_read_url"`
	KetoWriteURL           string  `validate:"required" envconfig:"keto_write_url"`
	ServiceName            string  `envconfig:"service_name" default:"authz-api"`
	SentryDSN              string  `envconfig:"sentry_dns"`
	SentryTracesSampleRate float64 `validate:"required_with=SentryDSN" envconfig:"sentry_traces_sample_rate"`
}

const (
	// Keep these hardcoded until the permission system stabilizes.
	publicPermissionAllowlistCSV     = "" // start with nothing allowlisted
	relationshipMutationAllowlistCSV = "" // start with nothing allowlisted
)

func main() {
	cfg := Config{}
	envconfig.Process("API", &cfg)

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Errorf("could not configure server: %w", err))
	}

	publicPermAllowlist, err := domain.ParsePermissionAllowlist(publicPermissionAllowlistCSV)
	if err != nil {
		panic(fmt.Errorf("invalid public permission allowlist: %w", err))
	}
	relMutationAllowlist, err := domain.ParseRelationshipMutationAllowlist(relationshipMutationAllowlistCSV)
	if err != nil {
		panic(fmt.Errorf("invalid relationship mutation allowlist: %w", err))
	}

	psql, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		panic(err)
	}

	kratosClient := kratosclient.NewClient(cfg.KratosURL)
	ketoAuthz := ketoclient.NewClient(cfg.KetoReadURL, cfg.KetoWriteURL)
	rolesSvc := commonroles.NewKetoService(ketoAuthz, "app", "tadoku")
	roleMgmt := commonroles.NewKetoManager(ketoAuthz, "app", "tadoku")
	postgresRepository := repository.NewRepository(psql)

	roleGet := domain.NewRoleGet(rolesSvc)
	roleUpdate := domain.NewRoleUpdate(kratosClient, postgresRepository, rolesSvc, roleMgmt)
	publicPermissionCheck := domain.NewPublicPermissionCheck(ketoAuthz, publicPermAllowlist)
	internalPermissionCheck := domain.NewInternalPermissionCheck(ketoAuthz)
	relationshipWriter := domain.NewRelationshipWriter(ketoAuthz, relMutationAllowlist)

	server := rest.NewServer(
		roleGet,
		roleUpdate,
		publicPermissionCheck,
		internalPermissionCheck,
		relationshipWriter,
	)

	e := echo.New()
	e.Use(middleware.Recover())

	// Health endpoints: allow K8s probes without auth, require admin if JWT is present
	optAuth := tadokumiddleware.OptionalAdminAuth(cfg.JWKS, rolesSvc)
	pgChecker := health.NewPostgresChecker("postgres", psql)
	e.GET("/livez", health.LivezHandler, optAuth)
	e.GET("/readyz", health.ReadyzHandler([]health.HealthChecker{pgChecker}), optAuth)

	// Business endpoints: full auth middleware stack
	api := e.Group("")
	api.Use(tadokumiddleware.Logger([]string{"/ping", "/internal/v1/ping"}))
	api.Use(tadokumiddleware.VerifyJWT(cfg.JWKS))
	api.Use(tadokumiddleware.Identity())
	api.Use(tadokumiddleware.RolesFromKeto(rolesSvc))
	api.Use(tadokumiddleware.RequireServiceAudience(cfg.ServiceName))

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: cfg.SentryTracesSampleRate,
		}); err != nil {
			panic(fmt.Errorf("sentry initialization failed: %v", err))
		}
		api.Use(sentryecho.New(sentryecho.Options{}))
	}

	openapi.RegisterHandlersWithBaseURL(api, server, "")
	internal := api.Group("", tadokumiddleware.RequireServiceIdentity())
	internalapi.RegisterHandlers(internal, server)

	go func() {
		fmt.Printf("authz-api is now available at: http://localhost:%d\n", cfg.Port)
		if err := e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
