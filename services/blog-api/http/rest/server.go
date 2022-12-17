package rest

import (
	"errors"
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
	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	res, err := s.pageCreateService.CreatePage(ctx.Request().Context(), &pagecreate.PageCreateRequest{
		ID:          req.Id,
		Slug:        req.Slug,
		Title:       req.Title,
		Html:        req.Html,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, pagecreate.ErrPageAlreadyExists) || errors.Is(err, pagecreate.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:    res.ID,
		Slug:  res.Slug,
		Title: res.Title,
		Html:  res.Html,
	})
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
