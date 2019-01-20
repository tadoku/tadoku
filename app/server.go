package app

import (
	"log"
	"net/http"
	"sync"

	"github.com/creasty/configo"

	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/interfaces/services"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	AutoConfigure() error
	Router() services.Router
	RDB() *infra.RDB
	SQLHandler() rdb.SQLHandler

	Repositories() *Repositories
	Interactors() *Interactors

	HealthService() services.HealthService
	SessionService() services.SessionService
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	Port                 string `envconfig:"app_port", valid:"required"`
	DatabaseURL          string `envconfig:"database_url" valid:"required"`
	DatabaseMaxIdleConns int    `envconfig:"database_max_idle_conns" valid:"required"`
	DatabaseMaxOpenConns int    `envconfig:"database_max_open_conns" valid:"required"`

	router struct {
		result services.Router
		once   sync.Once
	}

	rdb struct {
		result *infra.RDB
		once   sync.Once
	}

	sqlHandler struct {
		result rdb.SQLHandler
		once   sync.Once
	}

	repositories struct {
		result *Repositories
		once   sync.Once
	}

	interactors struct {
		result *Interactors
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
		holder.result = services.NewSessionService(d.Interactors().Session)
	})
	return holder.result
}

// ------------------------------
// Repositories
// ------------------------------

func (d *serverDependencies) Repositories() *Repositories {
	holder := &d.repositories
	holder.once.Do(func() {
		holder.result = NewRepositories(d.SQLHandler())
	})
	return holder.result
}

// ------------------------------
// Interactors
// ------------------------------

func (d *serverDependencies) Interactors() *Interactors {
	holder := &d.interactors
	holder.once.Do(func() {
		holder.result = NewInteractors(d.Repositories())
	})
	return holder.result
}

// ------------------------------
// Router
// ------------------------------

func (d *serverDependencies) Router() services.Router {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = infra.NewRouter(d.Port, d.routes()...)
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

// ------------------------------
// Relational database
// ------------------------------

func (d *serverDependencies) RDB() *infra.RDB {
	holder := &d.rdb
	holder.once.Do(func() {
		var err error
		holder.result, err = infra.NewRDB(d.DatabaseURL, d.DatabaseMaxIdleConns, d.DatabaseMaxOpenConns)

		if err != nil {
			// @TODO: we should handle errors more gracefully
			log.Fatalf("Failed to initialize connection pool with database: %v\n", err)
		}
	})
	return holder.result
}

func (d *serverDependencies) SQLHandler() rdb.SQLHandler {
	holder := &d.sqlHandler
	holder.once.Do(func() {
		holder.result = infra.NewSQLHandler(d.RDB())
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	router := d.Router()
	return router.StartListening()
}
