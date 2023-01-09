package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	contestCommandService contestcommand.Service,
	contestQueryService contestquery.Service,
	logQueryService logquery.Service,
) openapi.ServerInterface {
	return &Server{
		contestCommandService: contestCommandService,
		contestQueryService:   contestQueryService,
		logQueryService:       logQueryService,
	}
}

type Server struct {
	contestCommandService contestcommand.Service
	contestQueryService   contestquery.Service
	logQueryService       logquery.Service
}
