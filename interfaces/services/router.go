package services

// Router takes care of all the routing
type Router interface {
	StartListening() error
}

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(Context) error

// Route is a definition of a route
type Route struct {
	Method      string
	Path        string
	HandlerFunc HandlerFunc
	Restricted  bool
}
