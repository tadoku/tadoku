package query

import (
	"context"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/tadoku/tadoku/services/common/domain"
)

// userSearchSource implements fuzzy.Source for fuzzy matching on users
type userSearchSource struct {
	users []UserEntry
}

func (s userSearchSource) String(i int) string {
	u := s.users[i]
	return strings.ToLower(u.DisplayName + " " + u.Email)
}

func (s userSearchSource) Len() int {
	return len(s.users)
}

type ListUsersRequest struct {
	PerPage int64
	Page    int64
	Query   string
}

type ListUsersResponse struct {
	Users     []UserListEntry
	TotalSize int
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
	allUsers := s.userCache.GetUsers()

	var matchedUsers []UserEntry
	var totalSize int

	if req.Query == "" {
		// No search - paginate full list
		totalSize = len(allUsers)
		start := offset
		if start >= totalSize {
			matchedUsers = []UserEntry{}
		} else {
			end := start + perPage
			if end > totalSize {
				end = totalSize
			}
			matchedUsers = allUsers[start:end]
		}
	} else {
		// Fuzzy search
		source := userSearchSource{users: allUsers}
		matches := fuzzy.FindFrom(strings.ToLower(req.Query), source)
		totalSize = len(matches)

		start := offset
		if start >= totalSize {
			matchedUsers = []UserEntry{}
		} else {
			end := start + perPage
			if end > totalSize {
				end = totalSize
			}
			matchedUsers = make([]UserEntry, 0, end-start)
			for _, match := range matches[start:end] {
				matchedUsers = append(matchedUsers, allUsers[match.Index])
			}
		}
	}

	users := make([]UserListEntry, 0, len(matchedUsers))
	for _, u := range matchedUsers {
		users = append(users, UserListEntry{
			ID:          u.ID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			CreatedAt:   u.CreatedAt,
		})
	}

	return &ListUsersResponse{
		Users:     users,
		TotalSize: totalSize,
	}, nil
}
