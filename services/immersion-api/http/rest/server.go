package rest

import (
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	commandService command.Service,
	queryService query.Service,
	contestConfigurationOptions *domain.ContestConfigurationOptions,
) openapi.ServerInterface {
	return &Server{
		commandService:              commandService,
		queryService:                queryService,
		contestConfigurationOptions: contestConfigurationOptions,
	}
}

type Server struct {
	commandService command.Service
	queryService   query.Service

	// Service-per-function services
	contestConfigurationOptions *domain.ContestConfigurationOptions
}
