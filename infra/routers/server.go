package routers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/tadoku/api/domain"
)

// NewServerRouter instantiates an api router
func NewServerRouter() domain.Router {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e
}
