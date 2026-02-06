package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// COMMANDS

// Creates a new page
// (POST /pages/{namespace})
func (s *Server) PageCreate(ctx echo.Context, namespace string) error {
	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	resp, err := s.pageCreate.Execute(ctx.Request().Context(), &domain.PageCreateRequest{
		ID:          *req.Id,
		Namespace:   namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		HTML:        *req.Html,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPageAlreadyExists) || errors.Is(err, domain.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:          &resp.Page.ID,
		Slug:        resp.Page.Slug,
		Title:       resp.Page.Title,
		Html:        &resp.Page.HTML,
		PublishedAt: resp.Page.PublishedAt,
	})
}

// Updates an existing page
// (PUT /pages/{namespace}/{id})
func (s *Server) PageUpdate(ctx echo.Context, namespace string, id string) error {
	var req openapi.Page
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	resp, err := s.pageUpdate.Execute(ctx.Request().Context(), parsedID, &domain.PageUpdateRequest{
		Slug:        req.Slug,
		Namespace:   namespace,
		Title:       req.Title,
		HTML:        *req.Html,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPageAlreadyExists) || errors.Is(err, domain.ErrInvalidPage) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}
		if errors.Is(err, domain.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:          &resp.Page.ID,
		Slug:        resp.Page.Slug,
		Title:       resp.Page.Title,
		Html:        &resp.Page.HTML,
		PublishedAt: resp.Page.PublishedAt,
	})
}

// Deletes an existing page
// (DELETE /pages/{namespace}/{id})
func (s *Server) PageDelete(ctx echo.Context, namespace string, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = s.pageDelete.Execute(ctx.Request().Context(), parsedID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Lists all versions of a page
// (GET /pages/{namespace}/{id}/versions)
func (s *Server) PageVersionList(ctx echo.Context, namespace string, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	versions, err := s.pageVersionList.List(ctx.Request().Context(), parsedID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.PageVersions{
		Versions: make([]openapi.PageVersion, len(versions)),
	}
	for i, v := range versions {
		res.Versions[i] = openapi.PageVersion{
			Id:        v.ID,
			Version:   v.Version,
			Title:     v.Title,
			CreatedAt: v.CreatedAt,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

// Gets a specific version of a page
// (GET /pages/{namespace}/{id}/versions/{contentId})
func (s *Server) PageVersionGet(ctx echo.Context, namespace string, id string, contentId uuid.UUID) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	v, err := s.pageVersionList.Get(ctx.Request().Context(), parsedID, contentId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.PageVersion{
		Id:        v.ID,
		Version:   v.Version,
		Title:     v.Title,
		Html:      &v.HTML,
		CreatedAt: v.CreatedAt,
	})
}

// QUERIES

// Returns page content for a given slug
// (GET /pages/{namespace}/{slug})
func (s *Server) PageFindBySlug(ctx echo.Context, namespace string, slug string) error {
	resp, err := s.pageFind.Execute(ctx.Request().Context(), &domain.PageFindRequest{
		Slug:      slug,
		Namespace: namespace,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPageNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, domain.ErrRequestInvalid) {
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Page{
		Id:    &resp.Page.ID,
		Slug:  resp.Page.Slug,
		Title: resp.Page.Title,
		Html:  &resp.Page.HTML,
	})
}

// lists all pages
// (GET /pages/{namespace})
func (s *Server) PageList(ctx echo.Context, namespace string, params openapi.PageListParams) error {
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

	resp, err := s.pageList.Execute(ctx.Request().Context(), &domain.PageListRequest{
		Namespace:     namespace,
		PageSize:      pageSize,
		Page:          page,
		IncludeDrafts: includeDrafts,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if !errors.Is(err, domain.ErrPageNotFound) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	res := openapi.Pages{
		Pages:         []openapi.Page{},
		NextPageToken: resp.NextPageToken,
		TotalSize:     resp.TotalSize,
	}

	for _, p := range resp.Pages {
		p := p
		res.Pages = append(res.Pages, openapi.Page{
			Id:          &p.ID,
			Slug:        p.Slug,
			Title:       p.Title,
			Html:        &p.HTML,
			PublishedAt: p.PublishedAt,
			CreatedAt:   &p.CreatedAt,
			UpdatedAt:   &p.UpdatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, res)
}
