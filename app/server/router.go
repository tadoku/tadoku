package server

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/tadoku/api/domain"
)

// NewRouter instantiates an api router
func NewRouter() domain.Router {
	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	return e
}
