package profile

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sahilm/fuzzy"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Service struct {
	repository *Repository
	cache      *UserCache
	roles      *commonroles.KetoService
	identities *kratosclient.Client
}

func NewService(repository *Repository, cache *UserCache, roles *commonroles.KetoService, identities *kratosclient.Client) *Service {
	return &Service{
		repository: repository,
		cache:      cache,
		roles:      roles,
		identities: identities,
	}
}

func (s *Service) SynchronizeUser(ctx context.Context, user *identity.User, now time.Time) (uuid.UUID, error) {
	userID, err := user.UUID()
	if err != nil {
		return uuid.Nil, err
	}
	if err := s.repository.SynchronizeUser(ctx, userID, user.DisplayName, user.CreatedAt, now); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (s *Service) LockUser(ctx context.Context, userID uuid.UUID) error {
	state, err := s.repository.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	if state.DeletionLocked || state.Deleted {
		return ErrAccountDeletionInProgress
	}
	return nil
}

func (s *Service) DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	return s.repository.DisplayNames(ctx, ids)
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

func (s *Service) FindProfile(ctx context.Context, userID uuid.UUID) (*PublicProfile, error) {
	identity, err := s.identities.FetchIdentity(ctx, userID)
	if err != nil {
		return nil, err
	}
	if identity.GetSchemaId() != "user" {
		return nil, fmt.Errorf("unexpected schema %s", identity.GetSchemaId())
	}

	traits, err := decodeIdentityTraits(identity.GetTraits())
	if err != nil {
		return nil, err
	}

	return &PublicProfile{
		DisplayName: traits.DisplayName,
		CreatedAt:   identity.GetCreatedAt(),
	}, nil
}

func (s *Service) FetchAccountCreatedAt(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	identity, err := s.identities.FetchIdentity(ctx, userID)
	if errors.Is(err, kratosclient.ErrNotFound) {
		return time.Time{}, ErrIdentityNotFound
	}
	if err != nil {
		return time.Time{}, err
	}
	if identity.GetSchemaId() != "user" {
		return time.Time{}, fmt.Errorf("unexpected schema %s", identity.GetSchemaId())
	}

	return identity.GetCreatedAt(), nil
}
