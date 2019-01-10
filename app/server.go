package app

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	Router() *echo.Echo
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	router struct {
		result *echo.Echo
		once   sync.Once
	}
}

func (d *serverDependencies) Router() *echo.Echo {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = echo.New()
		holder.result.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	router := d.Router()
	router.Logger.Fatal(router.Start(":1234"))
	return nil
}
