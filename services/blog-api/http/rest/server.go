package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/blog-api/domain/pagecreate"
	"github.com/tadoku/tadoku/services/blog-api/http/rest/openapi"
)

// NewServer creates a new server conforming to the OpenAPI spec
func NewServer(
	pageCreateService pagecreate.Service,
) openapi.ServerInterface {
	return &server{
		pageCreateService: pageCreateService,
	}
}

type server struct {
	pageCreateService pagecreate.Service
}

// Creates a new page
// (POST /pages)
func (s *server) PageCreate(ctx echo.Context) error {
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		ctx.Echo().Logger.Error("could not process request: %w", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	req := &pagecreate.PageCreateRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		ctx.Echo().Logger.Error("could not process request: %w", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	res, err := s.pageCreateService.CreatePage(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, pagecreate.ErrPageAlreadyExists) || errors.Is(err, pagecreate.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: %w", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Updates an existing page
// (PUT /pages/{id})
func (s *server) PageUpdate(ctx echo.Context, id string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Returns page content for a given slug
// (GET /pages/{pageSlug})
func (s *server) PageFindBySlug(ctx echo.Context, pageSlug string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Checks if service is responsive
// (GET /ping)
func (s *server) Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}
