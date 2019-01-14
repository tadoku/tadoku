package infra

import (
	"log"
	"net/http"

	"github.com/labstack/echo"

	"github.com/tadoku/api/interfaces/services"
)

// NewRouter instantiates a router
func NewRouter(
	routes ...services.Route,
) services.Router {
	e := echo.New()

	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			e.GET(route.Path, wrap(route.HandlerFunc))
		case http.MethodPost:
			e.POST(route.Path, wrap(route.HandlerFunc))
		default:
			log.Fatalf("HTTP verb %v is not supported", route.Method)
		}
	}

	return e
}

func wrap(h services.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(c)
	}
}
