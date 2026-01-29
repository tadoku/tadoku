package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the tags used by a user, combined with default tags
// (GET /users/{user_id}/tags)
func (s *Server) GetUserTags(ctx echo.Context, userId types.UUID, params openapi.GetUserTagsParams) error {
	userID := userId

	pageSize := 50
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	page := 0
	if params.Page != nil {
		page = *params.Page
	}

	prefix := ""
	if params.Prefix != nil {
		prefix = *params.Prefix
	}

	res, err := s.queryService.FetchUserTags(ctx.Request().Context(), &query.FetchUserTagsRequest{
		UserID: userID,
		Prefix: prefix,
		Limit:  pageSize,
		Offset: page * pageSize,
	})
	if err != nil {
		if errors.Is(err, query.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	tags := make([]openapi.UserTagItem, len(res.Tags))
	for i, t := range res.Tags {
		tags[i] = openapi.UserTagItem{
			Tag:        t.Tag,
			UsageCount: int(t.UsageCount),
		}
	}

	return ctx.JSON(http.StatusOK, openapi.UserTags{
		Tags:        tags,
		DefaultTags: res.DefaultTags,
	})
}
