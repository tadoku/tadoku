package main

import (
	"database/sql"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/kelseyhightower/envconfig"
	"github.com/tadoku/tadoku/services/blog-api/http/rest"
	"github.com/tadoku/tadoku/services/blog-api/http/rest/openapi"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type Config struct {
	PostgresURL string `valid:"required" envconfig:"postgres_url"`
	Port        int64  `valid:"required"`
}

func main() {
	cfg := Config{}
	envconfig.Process("API", &cfg)

	valid, err := govalidator.ValidateStruct(cfg)
	if err != nil || !valid {
		panic(fmt.Errorf("could not configure server: %w", err))
	}

	_, err = sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(echomiddleware.Logger())

	server := rest.NewServer()

	openapi.RegisterHandlersWithBaseURL(e, server, "")

	fmt.Printf("blog-api is now available at: http://localhost:%d/v2\n", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
