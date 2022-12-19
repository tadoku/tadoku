package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/content-api/domain/postcommand"
	"github.com/tadoku/tadoku/services/content-api/domain/postquery"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// Creates a new post
// (POST /posts/{namespace})
func (s *Server) PostCreate(ctx echo.Context, namespace string) error {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return ctx.NoContent(http.StatusForbidden)
	}

	var req openapi.PostCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	post, err := s.postCommandService.CreatePost(ctx.Request().Context(), &postcommand.PostCreateRequest{
		ID:          *req.Id,
		Namespace:   namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		Content:     *req.Content,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, postcommand.ErrPostAlreadyExists) || errors.Is(err, postcommand.ErrInvalidPost) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:          &post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     &post.Content,
		PublishedAt: post.PublishedAt,
	})
}

// lists all posts
// (GET /posts/{namespace})
func (s *Server) PostList(ctx echo.Context, namespace string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Updates an existing post
// (PUT /posts/{namespace}/{id})
func (s *Server) PostUpdate(ctx echo.Context, namespace string, id string) error {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return ctx.NoContent(http.StatusForbidden)
	}

	var req openapi.PostUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	post, err := s.postCommandService.UpdatePost(ctx.Request().Context(), uuid.MustParse(id), &postcommand.PostUpdateRequest{
		Slug:        req.Slug,
		Namespace:   namespace,
		Title:       req.Title,
		Content:     *req.Content,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		if errors.Is(err, postcommand.ErrPostAlreadyExists) || errors.Is(err, postcommand.ErrInvalidPost) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:          &post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     &post.Content,
		PublishedAt: post.PublishedAt,
	})
}

// Returns page content for a given slug
// (GET /posts/{namespace}/{slug})
func (s *Server) PostFindBySlug(ctx echo.Context, namespace string, slug string) error {
	post, err := s.postQueryService.FindBySlug(ctx.Request().Context(), namespace, slug)
	if err != nil {
		if errors.Is(err, postquery.ErrPostNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Post{
		Id:      &post.ID,
		Slug:    post.Slug,
		Title:   post.Title,
		Content: &post.Content,
	})
}
