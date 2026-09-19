package profile

import (
	"context"
	"math"
	"strings"

	"github.com/sahilm/fuzzy"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type userCache interface {
	Users() []CachedUser
}

type roleProvider interface {
	ClaimsForSubjects(context.Context, []string) (map[string]commonroles.Claims, error)
}

type Service struct {
	cache       userCache
	roles       roleProvider
	permissions *permissions.Checker
}

func NewService(cache userCache, roles roleProvider, checker *permissions.Checker) *Service {
	return &Service{
		cache:       cache,
		roles:       roles,
		permissions: checker,
	}
}

func (s *Service) ListUsers(ctx context.Context, pageSize, page int, query string) (*UserList, error) {
	if err := s.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if page < 0 {
		page = 0
	}

	users := s.cache.Users()
	if query != "" {
		users = searchUsers(users, query)
	}
	totalSize := len(users)
	users = userPage(users, pageSize, page)

	subjectIDs := make([]string, 0, len(users))
	for _, user := range users {
		subjectIDs = append(subjectIDs, user.ID)
	}
	claims, err := s.roles.ClaimsForSubjects(ctx, subjectIDs)
	if err != nil {
		return nil, errx.NewUnavailableError("list user roles", err)
	}

	result := make([]User, 0, len(users))
	for _, user := range users {
		role := "user"
		if claims[user.ID].Admin {
			role = "admin"
		} else if claims[user.ID].Banned {
			role = "banned"
		}
		result = append(result, User{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			CreatedAt:   user.CreatedAt,
			Role:        role,
		})
	}

	return &UserList{Users: result, TotalSize: totalSize}, nil
}

type userSearchSource []CachedUser

func (s userSearchSource) String(i int) string {
	return strings.ToLower(s[i].DisplayName + " " + s[i].Email)
}

func (s userSearchSource) Len() int {
	return len(s)
}

func searchUsers(users []CachedUser, query string) []CachedUser {
	matches := fuzzy.FindFrom(strings.ToLower(query), userSearchSource(users))
	result := make([]CachedUser, 0, len(matches))
	for _, match := range matches {
		result = append(result, users[match.Index])
	}
	return result
}

func userPage(users []CachedUser, pageSize, page int) []CachedUser {
	offset := int64(page)
	if offset > math.MaxInt64/int64(pageSize) {
		return []CachedUser{}
	}
	offset *= int64(pageSize)
	if offset >= int64(len(users)) {
		return []CachedUser{}
	}

	start := int(offset)
	end := start + min(pageSize, len(users)-start)
	return users[start:end]
}
