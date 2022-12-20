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
// (POST /pages/{namespace})
func (s *Server) PageCreate(ctx echo.Context, namespace string) error {
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
		Namespace:   namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		Html:        *req.Html,
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
		Html:        &page.Html,
		PublishedAt: page.PublishedAt,
	})
}

// Updates an existing page
// (PUT /pages/{namespace}/{id})
func (s *Server) PageUpdate(ctx echo.Context, namespace string, id string) error {
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
		Namespace:   namespace,
		Title:       req.Title,
		Html:        *req.Html,
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
		Html:        &page.Html,
		PublishedAt: page.PublishedAt,
	})
}

// Returns page content for a given slug
// (GET /pages/{namespace}/{slug})
func (s *Server) PageFindBySlug(ctx echo.Context, namespace string, slug string) error {
	page, err := s.pageQueryService.FindBySlug(ctx.Request().Context(), &pagequery.PageFindRequest{
		Slug:      slug,
		Namespace: namespace,
	})
	if err != nil {
		if errors.Is(err, pagequery.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:    &page.ID,
		Slug:  page.Slug,
		Title: page.Title,
		Html:  &page.Html,
	})
}

// lists all pages
// (GET /pages/{namespace})
func (s *Server) PageList(ctx echo.Context, namespace string, params openapi.PageListParams) error {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return ctx.NoContent(http.StatusForbidden)
	}

	pageSize := 0
	page := 0
	includeDrafts := true

	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if params.Page != nil {
		page = *params.Page
	}
	if params.IncludeDrafts != nil {
		includeDrafts = *params.IncludeDrafts
	}

	list, err := s.pageQueryService.ListPages(ctx.Request().Context(), &pagequery.PageListRequest{
		Namespace:     namespace,
		PageSize:      pageSize,
		Page:          page,
		IncludeDrafts: includeDrafts,
	})
	if err != nil && !errors.Is(err, pagequery.ErrPageNotFound) {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.Pages{
		Pages:         []openapi.Page{},
		NextPageToken: list.NextPageToken,
		TotalSize:     list.TotalSize,
	}

	for _, page := range list.Pages {
		page := page
		res.Pages = append(res.Pages, openapi.Page{
			Id:          &page.ID,
			Slug:        page.Slug,
			Title:       page.Title,
			PublishedAt: page.PublishedAt,
			CreatedAt:   &page.CreatedAt,
			UpdatedAt:   &page.UpdatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, res)
}
