package app

import (
	"net/http"
	"sync"

	"github.com/creasty/configo"

	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/services"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	AutoConfigure() error
	Router() services.Router
	HealthService() services.HealthService
	SessionService() services.SessionService
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	DatabaseURL          string `envconfig:"database_url" valid:"required"`
	DatabaseMaxIdleConns string `envconfig:"database_max_idle_conns" valid:"required"`
	DatabaseMaxOpenConns string `envconfig:"database_max_open_conns" valid:"required"`

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

func (d *serverDependencies) AutoConfigure() error {
	return configo.Load(d, configo.Option{})
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
		holder.result = infra.NewRouter(d.routes()...)
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
