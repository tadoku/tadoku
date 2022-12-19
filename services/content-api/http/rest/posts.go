package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Creates a new post
// (POST /posts)
func (s *Server) PostCreate(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// lists all posts
// (GET /posts)
func (s *Server) PostList(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Updates an existing post
// (PUT /posts/{id})
func (s *Server) PostUpdate(ctx echo.Context, id string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Returns page content for a given slug
// (GET /posts/{slug})
func (s *Server) PostFindBySlug(ctx echo.Context, slug string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
