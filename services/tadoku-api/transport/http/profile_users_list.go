package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ProfileUsersList(
	ctx context.Context,
	request openapi.ProfileUsersListRequestObject,
) (openapi.ProfileUsersListResponseObject, error) {
	pageSize, page, query := 0, 0, ""
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	if request.Params.Query != nil {
		query = *request.Params.Query
	}

	result, err := s.application.ListUsers(ctx, pageSize, page, query)
	if err != nil {
		s.logOperationError(ctx, "list users", err)
		return nil, err
	}

	users := make([]openapi.ProfileUserListEntry, 0, len(result.Users))
	for _, user := range result.Users {
		entry := openapi.ProfileUserListEntry{
			Id:          user.ID,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
		}
		if user.Role != "user" {
			role := user.Role
			entry.Role = &role
		}
		users = append(users, entry)
	}

	return openapi.ProfileUsersList200JSONResponse{
		Users:     users,
		TotalSize: result.TotalSize,
	}, nil
}
