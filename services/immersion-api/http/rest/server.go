package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	contestCommandService contestcommand.Service,
) openapi.ServerInterface {
	return &Server{
		contestCommandService: contestCommandService,
	}
}

type Server struct {
	contestCommandService contestcommand.Service
}
