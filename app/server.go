package app

import (
	"sync"

	"github.com/tadoku/api/app/server"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	Router() server.Router
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	router struct {
		result server.Router
		once   sync.Once
	}
}

func (d *serverDependencies) Router() server.Router {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = server.NewRouter()
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	router := d.Router()
	router.Start(":1234")
	return nil
}
