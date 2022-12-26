package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer() openapi.ServerInterface {
	return &Server{}
}

type Server struct {
}
