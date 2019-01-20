package services

// Router takes care of all the routing
type Router interface {
	StartListening() error
}

// HandlerFunc of the router
type HandlerFunc func(Context) error

// Route is a definition of a route
type Route struct {
	Method      string
	Path        string
	HandlerFunc HandlerFunc
}
