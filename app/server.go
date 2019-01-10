package app

import (
	"net/http"
	"sync"

	"github.com/labstack/echo"
)

// Server is a dependency container for the api
type Server interface {
	Router() *echo.Echo
}

// NewServer instantiates a new api server
func NewServer() Server {
	return &server{}
}

type server struct {
	router struct {
		result *echo.Echo
		once   sync.Once
	}
}

func (s *server) Router() *echo.Echo {
	holder := &s.router
	holder.once.Do(func() {
		holder.result = echo.New()
		holder.result.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(s Server) error {
	router := s.Router()
	router.Logger.Fatal(router.Start(":1234"))
	return nil
}
