package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain/pagecommand"
	"github.com/tadoku/tadoku/services/content-api/domain/pagequery"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCommandService pagecommand.Service,
	pageQueryService pagequery.Service,
) openapi.ServerInterface {
	return &Server{
		pageCommandService: pageCommandService,
		pageQueryService:   pageQueryService,
	}
}

type Server struct {
	pageCommandService pagecommand.Service
	pageQueryService   pagequery.Service
}
