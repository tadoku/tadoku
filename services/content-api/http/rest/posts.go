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

// Creates a new post
// (POST /posts/{namespace})
func (s *Server) PostCreate(ctx echo.Context, namespace string) error {
	var req openapi.PostCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	resp, err := s.postCreate.Execute(ctx.Request().Context(), &domain.PostCreateRequest{
		ID:          *req.Id,
		Namespace:   namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		Content:     req.Content,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPostAlreadyExists) || errors.Is(err, domain.ErrInvalidPost) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:          &resp.Post.ID,
		Slug:        resp.Post.Slug,
		Title:       resp.Post.Title,
		Content:     resp.Post.Content,
		PublishedAt: resp.Post.PublishedAt,
	})
}

// Updates an existing post
// (PUT /posts/{namespace}/{id})
func (s *Server) PostUpdate(ctx echo.Context, namespace string, id string) error {
	var req openapi.PostUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	resp, err := s.postUpdate.Execute(ctx.Request().Context(), uuid.MustParse(id), &domain.PostUpdateRequest{
		Slug:        req.Slug,
		Namespace:   namespace,
		Title:       req.Title,
		Content:     req.Content,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPostAlreadyExists) || errors.Is(err, domain.ErrInvalidPost) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}
		if errors.Is(err, domain.ErrPostNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:          &resp.Post.ID,
		Slug:        resp.Post.Slug,
		Title:       resp.Post.Title,
		Content:     resp.Post.Content,
		PublishedAt: resp.Post.PublishedAt,
	})
}

// Deletes an existing post
// (DELETE /posts/{namespace}/{id})
func (s *Server) PostDelete(ctx echo.Context, namespace string, id string) error {
	err := s.postDelete.Execute(ctx.Request().Context(), uuid.MustParse(id))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrPostNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// QUERIES

// Returns page content for a given slug
// (GET /posts/{namespace}/{slug})
func (s *Server) PostFindBySlug(ctx echo.Context, namespace string, slug string) error {
	resp, err := s.postFind.Execute(ctx.Request().Context(), &domain.PostFindRequest{
		Namespace: namespace,
		Slug:      slug,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, domain.ErrRequestInvalid) {
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:          &resp.Post.ID,
		Slug:        resp.Post.Slug,
		Title:       resp.Post.Title,
		Content:     resp.Post.Content,
		PublishedAt: resp.Post.PublishedAt,
	})
}

// lists all posts
// (GET /posts/{namespace})
func (s *Server) PostList(ctx echo.Context, namespace string, params openapi.PostListParams) error {
	pageSize := 0
	page := 0
	includeDrafts := false

	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if params.Page != nil {
		page = *params.Page
	}
	if params.IncludeDrafts != nil {
		includeDrafts = *params.IncludeDrafts
	}

	resp, err := s.postList.Execute(ctx.Request().Context(), &domain.PostListRequest{
		Namespace:     namespace,
		PageSize:      pageSize,
		Page:          page,
		IncludeDrafts: includeDrafts,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if !errors.Is(err, domain.ErrPostNotFound) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	res := openapi.Posts{
		Posts:         []openapi.Post{},
		NextPageToken: resp.NextPageToken,
		TotalSize:     resp.TotalSize,
	}

	for _, p := range resp.Posts {
		p := p
		res.Posts = append(res.Posts, openapi.Post{
			Id:          &p.ID,
			Slug:        p.Slug,
			Title:       p.Title,
			Content:     p.Content,
			PublishedAt: p.PublishedAt,
			CreatedAt:   &p.CreatedAt,
			UpdatedAt:   &p.UpdatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, res)
}
