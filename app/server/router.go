package server

import (
	"net/http"

	"github.com/labstack/echo"
)

// Router takes care of all the routing for the api
type Router interface {
	Start(address string) error
}

// NewRouter instantiates an api router
func NewRouter() Router {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	return e
}
