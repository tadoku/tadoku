package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/blog-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer() openapi.ServerInterface {
	return &server{}
}

type server struct{}

// Creates a new page
// (POST /pages)
func (s *server) CreatePage(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Updates an existing page
// (PUT /pages/{id})
func (s *server) UpdatePage(ctx echo.Context, id string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Returns page content for a given slug
// (GET /pages/{pageSlug})
func (s *server) FindPageBySlug(ctx echo.Context, pageSlug string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
