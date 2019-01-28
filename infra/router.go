package infra

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/tadoku/api/interfaces/services"
)

var restricted echo.MiddlewareFunc

// NewRouter instantiates a router
func NewRouter(
	port string,
	jwtSecret string,
	routes ...services.Route,
) services.Router {
	e := echo.New()
	restricted = newJWTMiddleware(jwtSecret)

	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			e.GET(route.Path, wrap(route))
		case http.MethodPost:
			e.POST(route.Path, wrap(route))
		default:
			log.Fatalf("HTTP verb %v is not supported", route.Method)
		}
	}

	return router{e, port}
}

func newJWTMiddleware(secret string) echo.MiddlewareFunc {
	cfg := middleware.JWTConfig{
		Claims:     &jwtClaims{},
		SigningKey: []byte(secret),
	}
	return middleware.JWTWithConfig(cfg)
}

func wrap(r services.Route) echo.HandlerFunc {
	handler := func(c echo.Context) error {
		return r.HandlerFunc(&context{c})
	}

	if r.Restricted {
		handler = restricted(handler)
	}

	return handler
}

type router struct {
	*echo.Echo
	port string
}

func (r router) StartListening() error {
	return r.Start(":" + r.port)
}
