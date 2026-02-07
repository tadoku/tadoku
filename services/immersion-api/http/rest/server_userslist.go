package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Lists all users (admin only)
// (GET /users)
func (s *Server) UsersList(ctx echo.Context, params openapi.UsersListParams) error {
	perPage := int64(20)
	page := int64(0)
	queryStr := ""

	if params.PageSize != nil {
		perPage = int64(*params.PageSize)
	}
	if params.Page != nil {
		page = int64(*params.Page)
	}
	if params.Query != nil {
		queryStr = *params.Query
	}

	result, err := s.userList.Execute(ctx.Request().Context(), &domain.UserListRequest{
		PerPage: perPage,
		Page:    page,
		Query:   queryStr,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrAuthzUnavailable) {
			return ctx.NoContent(http.StatusServiceUnavailable)
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	users := make([]openapi.UserListEntry, len(result.Users))
	for i, u := range result.Users {
		entry := openapi.UserListEntry{
			Id:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
		}
		// Only set role if it's not the default "user" role
		if u.Role != "" {
			role := u.Role
			entry.Role = &role
		}
		users[i] = entry
	}

	return ctx.JSON(http.StatusOK, openapi.UserList{
		Users:     users,
		TotalSize: result.TotalSize,
	})
}
