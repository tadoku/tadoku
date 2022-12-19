package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/content-api/domain/pagecommand"
	"github.com/tadoku/tadoku/services/content-api/domain/pagequery"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// Creates a new page
// (POST /pages)
func (s *Server) PageCreate(ctx echo.Context) error {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return ctx.NoContent(http.StatusForbidden)
	}

	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	page, err := s.pageCommandService.CreatePage(ctx.Request().Context(), &pagecommand.PageCreateRequest{
		ID:          *req.Id,
		Slug:        req.Slug,
		Title:       req.Title,
		Html:        req.Html,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, pagecommand.ErrPageAlreadyExists) || errors.Is(err, pagecommand.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:          &page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: page.PublishedAt,
	})
}

// Updates an existing page
// (PUT /pages/{id})
func (s *Server) PageUpdate(ctx echo.Context, id string) error {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return ctx.NoContent(http.StatusForbidden)
	}

	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	page, err := s.pageCommandService.UpdatePage(ctx.Request().Context(), uuid.MustParse(id), &pagecommand.PageUpdateRequest{
		Slug:        req.Slug,
		Title:       req.Title,
		Html:        req.Html,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, pagecommand.ErrPageAlreadyExists) || errors.Is(err, pagecommand.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:          &page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: page.PublishedAt,
	})
}

// Returns page content for a given slug
// (GET /pages/{pageSlug})
func (s *Server) PageFindBySlug(ctx echo.Context, pageSlug string) error {
	page, err := s.pageQueryService.FindBySlug(ctx.Request().Context(), pageSlug)
	if err != nil {
		if errors.Is(err, pagequery.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:    &page.ID,
		Slug:  page.Slug,
		Title: page.Title,
		Html:  page.Html,
	})
}
