package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain/pagecommand"
	"github.com/tadoku/tadoku/services/content-api/domain/pagefind"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCommandService pagecommand.Service,
	pageFindService pagefind.Service,
) openapi.ServerInterface {
	return &Server{
		pageCommandService: pageCommandService,
		pageFindService:    pageFindService,
	}
}

type Server struct {
	pageCommandService pagecommand.Service
	pageFindService    pagefind.Service
}
