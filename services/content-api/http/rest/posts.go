package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/content-api/domain/postcommand"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
)

// Creates a new post
// (POST /posts)
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
// (GET /posts)
func (s *Server) PostList(ctx echo.Context, namespace string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Updates an existing post
// (PUT /posts/{id})
func (s *Server) PostUpdate(ctx echo.Context, namespace string, id string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Returns page content for a given slug
// (GET /posts/{slug})
func (s *Server) PostFindBySlug(ctx echo.Context, namespace string, slug string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}
