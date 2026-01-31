package query

import (
	"context"

	"github.com/tadoku/tadoku/services/common/domain"
)

type ListUsersRequest struct {
	PerPage int64
	Page    int64
	Query   string
}

type ListUsersResponse struct {
	Users         []UserListEntry
	NextPageToken string
}

type UserListEntry struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

func (s *ServiceImpl) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	if session.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}

	perPage := int(req.PerPage)
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	page := int(req.Page)
	if page < 0 {
		page = 0
	}

	offset := page * perPage
	cacheUsers, hasMore := s.userCache.Search(req.Query, perPage, offset)

	users := make([]UserListEntry, 0, len(cacheUsers))
	for _, u := range cacheUsers {
		users = append(users, UserListEntry{
			ID:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
		})
	}

	nextPageToken := ""
	if hasMore {
		nextPageToken = "has_more"
	}

	return &ListUsersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
	}, nil
}
