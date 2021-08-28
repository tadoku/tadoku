package app

import (
	"log"
	"net/http"
	"sync"

	"github.com/creasty/configo"

	"github.com/tadoku/tadoku/services/reading-contest-api/domain"
	"github.com/tadoku/tadoku/services/reading-contest-api/infra"
	"github.com/tadoku/tadoku/services/reading-contest-api/interfaces/rdb"
	"github.com/tadoku/tadoku/services/reading-contest-api/interfaces/services"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	AutoConfigure() error

	Init()
	Router() services.Router
	ErrorReporter() usecases.ErrorReporter

	RDB() *infra.RDB
	SQLHandler() rdb.SQLHandler

	Repositories() *Repositories
	Interactors() *Interactors
	Services() *Services
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	Environment          domain.Environment `envconfig:"app_env" valid:"environment" default:"development"`
	DatabaseURL          string             `envconfig:"database_url" valid:"required"`
	DatabaseMaxIdleConns int                `envconfig:"database_max_idle_conns" valid:"required"`
	DatabaseMaxOpenConns int                `envconfig:"database_max_open_conns" valid:"required"`
	CORSAllowedOrigins   []string           `envconfig:"cors_allowed_origins" valid:"required"`
	ErrorReporterDSN     string             `envconfig:"error_reporter_dsn"`
	JWTSecret            string             `envconfig:"jwt_secret" valid:"required"`
	Port                 string             `envconfig:"app_port" valid:"required"`
	SessionCookieName    string             `envconfig:"user_session_cookie_name" valid:"required"`
	TimeZone             string             `envconfig:"app_timezone" valid:"required"`

	router struct {
		result services.Router
		once   sync.Once
	}

	errorReporter struct {
		result usecases.ErrorReporter
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

	services struct {
		result *Services
		once   sync.Once
	}
}

func (d *serverDependencies) AutoConfigure() error {
	infra.ConfigureCustomValidators()
	return configo.Load(d, configo.Option{})
}

// ------------------------------
// Services
// ------------------------------

func (d *serverDependencies) Services() *Services {
	holder := &d.services
	holder.once.Do(func() {
		holder.result = NewServices(d.Interactors())
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
		holder.result = infra.NewRouter(d.Environment, d.Port, d.JWTSecret, d.SessionCookieName, d.CORSAllowedOrigins, d.ErrorReporter(), d.routes()...)
	})
	return holder.result
}

func (d *serverDependencies) routes() []services.Route {
	routes := []services.Route{
		// Service infra
		{Method: http.MethodGet, Path: "/ping", HandlerFunc: d.Services().Health.Ping},

		// Contests
		{Method: http.MethodGet, Path: "/contests", HandlerFunc: d.Services().Contest.All},
		{Method: http.MethodGet, Path: "/contests/:id", HandlerFunc: d.Services().Contest.Get},
		{Method: http.MethodPost, Path: "/contests", HandlerFunc: d.Services().Contest.Create, MinRole: domain.RoleAdmin},
		{Method: http.MethodPut, Path: "/contests/:id", HandlerFunc: d.Services().Contest.Update, MinRole: domain.RoleAdmin},

		// Rankings
		// TODO: Rename Get to All
		{Method: http.MethodGet, Path: "/rankings", HandlerFunc: d.Services().Ranking.Get},

		// Ranking registrations
		{Method: http.MethodPost, Path: "/ranking_registrations", HandlerFunc: d.Services().Ranking.Create, MinRole: domain.RoleUser},
		{Method: http.MethodGet, Path: "/ranking_registrations", HandlerFunc: d.Services().Ranking.RankingsForRegistration},
		{Method: http.MethodGet, Path: "/ranking_registrations/:id/current", HandlerFunc: d.Services().Ranking.CurrentRegistration, MinRole: domain.RoleUser},

		// Contest logs
		{Method: http.MethodPost, Path: "/contest_logs", HandlerFunc: d.Services().ContestLog.Create, MinRole: domain.RoleUser},
		// TODO: Rename Get to All
		{Method: http.MethodGet, Path: "/contest_logs", HandlerFunc: d.Services().ContestLog.Get},
		{Method: http.MethodPut, Path: "/contest_logs/:id", HandlerFunc: d.Services().ContestLog.Update, MinRole: domain.RoleUser},
		{Method: http.MethodDelete, Path: "/contest_logs/:id", HandlerFunc: d.Services().ContestLog.Delete, MinRole: domain.RoleUser},
	}

	return routes
}

func (d *serverDependencies) ErrorReporter() usecases.ErrorReporter {
	holder := &d.errorReporter
	holder.once.Do(func() {
		var err error
		holder.result, err = infra.NewErrorReporter(d.ErrorReporterDSN)

		if err != nil {
			log.Fatalf("failed to initialize error reporter: %v\n", err)
		}
	})
	return holder.result
}

func (d *serverDependencies) Init() {
	_ = d.ErrorReporter()
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
			log.Fatalf("failed to initialize connection pool with database: %v\n", err)
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
	d.Init()

	router := d.Router()
	return router.StartListening()
}