package profile

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sahilm/fuzzy"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Service struct {
	repository *Repository
	cache      *UserCache
	roles      *commonroles.KetoService
}

func NewService(repository *Repository, cache *UserCache, roles *commonroles.KetoService) *Service {
	return &Service{
		repository: repository,
		cache:      cache,
		roles:      roles,
	}
}

func (s *Service) SignedUser(ctx context.Context) (SignedUser, error) {
	requestUser := identity.FromContext(ctx)
	if requestUser == nil {
		return SignedUser{}, ErrInvalidSignedUser
	}
	userID, err := uuid.Parse(requestUser.Subject)
	if err != nil {
		return SignedUser{}, ErrInvalidSignedUser
	}
	return SignedUser{
		ID:               userID,
		DisplayName:      requestUser.DisplayName,
		sessionCreatedAt: requestUser.CreatedAt,
	}, nil
}

func (s *Service) SynchronizeUser(ctx context.Context, user SignedUser, now time.Time) error {
	return s.repository.SynchronizeUser(ctx, user, now)
}

func (s *Service) LockUser(ctx context.Context, userID uuid.UUID) error {
	return s.repository.LockUser(ctx, userID)
}

func (s *Service) ListUsers(ctx context.Context, pageSize, page int, query string) (*UserList, error) {
	pageSize, page = normalizeUserPage(pageSize, page)

	users, err := s.cache.Users(ctx)
	if err != nil {
		return nil, errx.NewUnavailableError("list users", err)
	}
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

func normalizeUserPage(pageSize, page int) (int, int) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if page < 0 {
		page = 0
	}
	return pageSize, page
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
