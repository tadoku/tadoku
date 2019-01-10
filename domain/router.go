package domain

// Router takes care of all the routing
type Router interface {
	Start(address string) error
}

// HandlerFunc of the router
type HandlerFunc func(Context) error
