package domain

// Router takes care of all the routing
type Router interface {
	Start(address string) error
}
