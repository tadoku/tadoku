package main

import (
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/common/domain"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"
	"github.com/tadoku/tadoku/services/common/storage/memory"
	"github.com/tadoku/tadoku/services/immersion-api/client/ory"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/profilequery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
)

type Config struct {
	PostgresURL string `validate:"required" envconfig:"postgres_url"`
	Port        int64  `validate:"required"`
	JWKS        string `validate:"required"`
	KratosURL   string `validate:"required" envconfig:"kratos_url"`
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

	contestRepository := postgres.NewContestRepository(psql)
	logRepository := postgres.NewLogRepository(psql)
	roleRepository := memory.NewRoleRepository("/etc/tadoku/permissions/roles.yaml")

	e := echo.New()
	e.Use(tadokumiddleware.Logger([]string{"/ping"}))
	e.Use(tadokumiddleware.SessionJWT(cfg.JWKS))
	e.Use(tadokumiddleware.Session(roleRepository))

	clock, err := domain.NewClock("UTC")
	if err != nil {
		panic(err)
	}

	contestCommandService := contestcommand.NewService(contestRepository, clock)
	contestQueryService := contestquery.NewService(contestRepository, clock)
	logCommandService := logcommand.NewService(
		logRepository,
		contestRepository,
		clock,
	)
	logQueryService := logquery.NewService(logRepository, clock)
	profileQueryService := profilequery.NewService(contestRepository)

	server := rest.NewServer(
		contestCommandService,
		contestQueryService,
		logCommandService,
		logQueryService,
		profileQueryService,
	)

	openapi.RegisterHandlersWithBaseURL(e, server, "")

	fmt.Printf("immersion-api is now available at: http://localhost:%d/v2\n", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
