package app

import (
	"net/http"
	"sync"

	"github.com/tadoku/api/infra/router"
	"github.com/tadoku/api/interfaces/services"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	Router() services.Router
	HealthService() services.HealthService
	SessionService() services.SessionService
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	router struct {
		result services.Router
		once   sync.Once
	}

	healthService struct {
		result services.HealthService
		once   sync.Once
	}

	sessionService struct {
		result services.SessionService
		once   sync.Once
	}
}

// ------------------------------
// Services
// ------------------------------

func (d *serverDependencies) HealthService() services.HealthService {
	holder := &d.healthService
	holder.once.Do(func() {
		holder.result = services.NewHealthService()
	})
	return holder.result
}

func (d *serverDependencies) SessionService() services.SessionService {
	holder := &d.sessionService
	holder.once.Do(func() {
		holder.result = services.NewSessionService()
	})
	return holder.result
}

// ------------------------------
// Router
// ------------------------------

func (d *serverDependencies) Router() services.Router {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = router.NewRouter(d.routes()...)
	})
	return holder.result
}

func (d *serverDependencies) routes() []services.Route {
	return []services.Route{
		{Method: http.MethodGet, Path: "/ping", HandlerFunc: d.HealthService().Ping},
		{Method: http.MethodPost, Path: "/login", HandlerFunc: d.SessionService().Login},
		{Method: http.MethodPost, Path: "/register", HandlerFunc: d.SessionService().Register},
	}
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	router := d.Router()
	router.Start(":1234")
	return nil
}
