package app

// Server is a dependency container for the api
type Server interface {
}

// NewServer instantiates a new api server
func NewServer() Server {
	return &server{}
}

type server struct {
}

// RunServer starts the actual API server
func RunServer(s Server) error {
	return nil
}
