package query

import (
	"context"
	"strings"

	"github.com/tadoku/tadoku/services/common/domain"
)

type ListUsersRequest struct {
	PerPage int64
	Page    int64
	Email   string
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

	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	page := req.Page
	if page < 0 {
		page = 0
	}

	result, err := s.kratos.ListIdentities(ctx, perPage, page)
	if err != nil {
		return nil, err
	}

	users := make([]UserListEntry, 0, len(result.Identities))
	for _, identity := range result.Identities {
		// Filter by email if specified (case-insensitive contains)
		if req.Email != "" && !strings.Contains(strings.ToLower(identity.Email), strings.ToLower(req.Email)) {
			continue
		}

		users = append(users, UserListEntry{
			ID:          identity.ID,
			DisplayName: identity.DisplayName,
			Email:       identity.Email,
			CreatedAt:   identity.CreatedAt,
		})
	}

	nextPageToken := ""
	if result.HasMore {
		nextPageToken = "has_more"
	}

	return &ListUsersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
	}, nil
}
