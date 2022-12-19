package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain/pagecreate"
	"github.com/tadoku/tadoku/services/content-api/domain/pagefind"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCreateService pagecreate.Service,
	pageFindService pagefind.Service,
) openapi.ServerInterface {
	return &Server{
		pageCreateService: pageCreateService,
		pageFindService:   pageFindService,
	}
}

type Server struct {
	pageCreateService pagecreate.Service
	pageFindService   pagefind.Service
}
