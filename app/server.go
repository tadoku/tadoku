package app

import (
	"sync"

	"github.com/tadoku/api/app/server"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/services"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	Router() domain.Router
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	router struct {
		result domain.Router
		once   sync.Once
	}
}

func (d *serverDependencies) Router() domain.Router {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = server.NewRouter(
			services.NewHealthService(),
			services.NewSessionService(),
		)
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	router := d.Router()
	router.Start(":1234")
	return nil
}
