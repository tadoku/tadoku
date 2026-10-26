package profile

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Service struct {
	repository *Repository
	identities *identity.ReadService
}

// RoleLookup returns each user's role name in the context tenant: admin, banned or user.
type RoleLookup func(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)

func NewService(repository *Repository, identities *identity.ReadService) *Service {
	return &Service{repository: repository, identities: identities}
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

// Within a transaction, ErrAccountDeletionInProgress still leaves the user row
// locked; callers may finish read-only work but must not mutate that user's data.
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

// LockForDeletion requires the caller's transaction. From its commit the user can no longer sign in or change data;
// a user whose deletion is already requested or done returns ErrAccountDeletionInProgress.
func (s *Service) LockForDeletion(ctx context.Context, userID uuid.UUID, now time.Time) error {
	state, err := s.repository.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	if state.DeletionLocked || state.Deleted {
		return ErrAccountDeletionInProgress
	}
	return s.repository.MarkDeletionLocked(ctx, userID, now)
}

// Tombstone requires the caller's transaction and locks the user row before any other row of the deletion. It returns
// false without changes for a user who is already deleted, and ErrDeletionNotRequested for a user who is not locked
// for deletion.
func (s *Service) Tombstone(ctx context.Context, userID uuid.UUID, now time.Time) (bool, error) {
	state, err := s.repository.LockUser(ctx, userID)
	if err != nil {
		return false, err
	}
	if state.Deleted {
		return false, nil
	}
	if !state.DeletionLocked {
		return false, ErrDeletionNotRequested
	}
	return true, s.repository.Tombstone(ctx, userID, now)
}

func (s *Service) DisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	return s.repository.DisplayNames(ctx, ids)
}

func (s *Service) ListUsers(
	ctx context.Context,
	pageSize, page int,
	query string,
	roles RoleLookup,
) (*UserList, error) {
	pageSize, page = normalizeUserPage(pageSize, page)

	accounts, err := s.identities.ListAllAccounts(ctx, query)
	if err != nil {
		return nil, err
	}

	totalSize := len(accounts)
	users := userPage(accounts, pageSize, page)

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	userRoles, err := roles(ctx, userIDs)
	if err != nil {
		return nil, errx.NewUnavailableError("list user roles", err)
	}

	result := make([]User, 0, len(users))
	for _, user := range users {
		result = append(result, User{
			ID:          user.ID.String(),
			DisplayName: user.DisplayName.String(),
			Email:       user.Email.String(),
			CreatedAt:   formatAccountCreatedAt(user.CreatedAt),
			Role:        userRoles[user.ID],
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
	return value.UTC().Format("2006-01-02T15:04:05Z")
}
