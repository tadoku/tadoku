package profile

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Service struct {
	repository *Repository
	roles      *permissions.KetoService
	identities *identity.ReadService
}

func NewService(
	repository *Repository,
	roles *permissions.KetoService,
	identities *identity.ReadService,
) *Service {
	return &Service{
		repository: repository,
		roles:      roles,
		identities: identities,
	}
}

func (s *Service) SynchronizeUser(ctx context.Context, user *requestidentity.User, now time.Time) (uuid.UUID, error) {
	userID, ok := requestidentity.ActorID(ctx)
	if !ok {
		return uuid.Nil, errx.NewInternalError("invalid signed user identity")
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

	accounts, err := s.identities.ListAllAccounts(ctx, query)
	if err != nil {
		return nil, err
	}

	totalSize := len(accounts)
	users := userPage(accounts, pageSize, page)

	subjectIDs := make([]string, 0, len(users))
	for _, user := range users {
		subjectIDs = append(subjectIDs, user.ID.String())
	}
	claims, err := s.roles.RolesForSubjects(ctx, subjectIDs)
	if err != nil {
		return nil, errx.NewUnavailableError("list user roles", err)
	}

	result := make([]User, 0, len(users))
	for _, user := range users {
		role := "user"
		if claims[user.ID.String()].Admin {
			role = "admin"
		} else if claims[user.ID.String()].Banned {
			role = "banned"
		}
		result = append(result, User{
			ID:          user.ID.String(),
			DisplayName: user.DisplayName.String(),
			Email:       user.Email.String(),
			CreatedAt:   formatAccountCreatedAt(user.CreatedAt),
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

func userPage(users []identity.Account, pageSize, page int) []identity.Account {
	offset := int64(page)
	if offset > math.MaxInt64/int64(pageSize) {
		return []identity.Account{}
	}
	offset *= int64(pageSize)
	if offset >= int64(len(users)) {
		return []identity.Account{}
	}

	start := int(offset)
	end := start + min(pageSize, len(users)-start)
	return users[start:end]
}

func (s *Service) FindProfile(ctx context.Context, userID uuid.UUID) (*PublicProfile, error) {
	account, err := s.identities.FindAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &PublicProfile{DisplayName: account.DisplayName.String(), CreatedAt: account.CreatedAt}, nil
}

func (s *Service) FetchAccountCreatedAt(ctx context.Context, userID uuid.UUID) (time.Time, error) {
	account, err := s.identities.FindAccount(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}
	return account.CreatedAt, nil
}

func formatAccountCreatedAt(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02T15:04:05Z")
}
