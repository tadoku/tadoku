package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
)

// Router takes care of all the routing
type Router interface {
	StartListening() error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(Context) error

// Route is a definition of a route
type Route struct {
	Method      string
	Path        string
	HandlerFunc HandlerFunc
	MinRole     domain.Role
}
