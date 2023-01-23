package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a profile of a user
// (GET /users/{userId}/profile)
func (s *Server) ProfileFindByUserID(ctx echo.Context, userId types.UUID) error {
	profile, err := s.queryService.FetchUserProfile(ctx.Request().Context(), userId)
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Errorf("could not fetch profile: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := &openapi.UserProfile{
		Id:          userId,
		DisplayName: profile.DisplayName,
		CreatedAt:   profile.CreatedAt,
	}

	return ctx.JSON(http.StatusOK, res)
}
