package rest

import (
	"github.com/tadoku/tadoku/services/content-api/domain/pagecommand"
	"github.com/tadoku/tadoku/services/content-api/domain/pagequery"
	"github.com/tadoku/tadoku/services/content-api/domain/postcommand"
	"github.com/tadoku/tadoku/services/content-api/domain/postquery"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCommandService pagecommand.Service,
	postCommandService postcommand.Service,
	pageQueryService pagequery.Service,
	postQueryService postquery.Service,
) openapi.ServerInterface {
	return &Server{
		pageCommandService: pageCommandService,
		postCommandService: postCommandService,
		pageQueryService:   pageQueryService,
		postQueryService:   postQueryService,
	}
}

type Server struct {
	pageCommandService pagecommand.Service
	postCommandService postcommand.Service

	pageQueryService pagequery.Service
	postQueryService postquery.Service
}
