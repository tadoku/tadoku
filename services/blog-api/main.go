package main

import (
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/blog-api/domain/pagecreate"
	"github.com/tadoku/tadoku/services/blog-api/domain/pagefind"
	"github.com/tadoku/tadoku/services/blog-api/http/rest"
	"github.com/tadoku/tadoku/services/blog-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/blog-api/storage/postgres"
	tadokumiddleware "github.com/tadoku/tadoku/services/common/middleware"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
)

type Config struct {
	PostgresURL string `validate:"required" envconfig:"postgres_url"`
	Port        int64  `validate:"required"`
	JWKS        string `validate:"required"`
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

	e := echo.New()
	e.Use(tadokumiddleware.Logger([]string{"/ping"}))
	e.Use(tadokumiddleware.SessionJWT(cfg.JWKS))
	e.Use(tadokumiddleware.Session())

	pageRepository := postgres.NewPageRepository(psql)

	pageCreateService := pagecreate.NewService(pageRepository)
	pageFindService := pagefind.NewService(pageRepository)

	server := rest.NewServer(
		pageCreateService,
		pageFindService,
	)

	openapi.RegisterHandlersWithBaseURL(e, server, "")

	fmt.Printf("blog-api is now available at: http://localhost:%d/v2\n", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
