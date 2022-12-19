package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/content-api/domain/pagecreate"
	"github.com/tadoku/tadoku/services/content-api/domain/pagefind"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// Creates a new page
// (POST /pages)
func (s *Server) PageCreate(ctx echo.Context) error {
	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	page, err := s.pageCreateService.CreatePage(ctx.Request().Context(), &pagecreate.PageCreateRequest{
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
		Id:    page.ID,
		Slug:  page.Slug,
		Title: page.Title,
		Html:  page.Html,
	})
}

// Updates an existing page
// (PUT /pages/{id})
func (s *Server) PageUpdate(ctx echo.Context, id string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Returns page content for a given slug
// (GET /pages/{pageSlug})
func (s *Server) PageFindBySlug(ctx echo.Context, pageSlug string) error {
	// ctx.Echo().Logger.Error("wtf", ctx.Get("user"))

	page, err := s.pageFindService.FindBySlug(ctx.Request().Context(), pageSlug)
	if err != nil {
		if errors.Is(err, pagefind.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:    page.ID,
		Slug:  page.Slug,
		Title: page.Title,
		Html:  page.Html,
	})
}
