package server

import (
	"github.com/labstack/echo"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/domain/services"
)

// NewRouter instantiates an api router
func NewRouter(
	healthService services.HealthService,
	sessionService services.SessionService,
) domain.Router {
	e := echo.New()

	e.GET("/ping", wrap(healthService.Ping))

	e.POST("/login", wrap(sessionService.Login))
	e.POST("/register", wrap(sessionService.Register))

	return e
}

func wrap(h domain.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(c)
	}
}
