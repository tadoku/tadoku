package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Lists all users (admin only)
// (GET /users)
func (s *Server) UsersList(ctx echo.Context, params openapi.UsersListParams) error {
	perPage := int64(20)
	page := int64(0)
	email := ""

	if params.PageSize != nil {
		perPage = int64(*params.PageSize)
	}
	if params.Page != nil {
		page = int64(*params.Page)
	}
	if params.Email != nil {
		email = *params.Email
	}

	result, err := s.queryService.ListUsers(ctx.Request().Context(), &query.ListUsersRequest{
		PerPage: perPage,
		Page:    page,
		Email:   email,
	})
	if err != nil {
		if errors.Is(err, query.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, query.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	users := make([]openapi.UserListEntry, len(result.Users))
	for i, u := range result.Users {
		users[i] = openapi.UserListEntry{
			Id:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
		}
	}

	return ctx.JSON(http.StatusOK, openapi.UserList{
		Users:         users,
		NextPageToken: result.NextPageToken,
	})
}
