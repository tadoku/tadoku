package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

var UserCreateContestYearlyLimit int32 = 12

type ContestPermissionCheckRepository interface {
	GetContestsByUserCountForYear(ctx context.Context, now time.Time, userID uuid.UUID) (int32, error)
}

type ContestPermissionCheckKratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
}

type ContestPermissionCheck struct {
	repo   ContestPermissionCheckRepository
	kratos ContestPermissionCheckKratosClient
	clock  commondomain.Clock
}

func NewContestPermissionCheck(repo ContestPermissionCheckRepository, kratos ContestPermissionCheckKratosClient, clock commondomain.Clock) *ContestPermissionCheck {
	return &ContestPermissionCheck{repo: repo, kratos: kratos, clock: clock}
}

func (s *ContestPermissionCheck) Execute(ctx context.Context) error {
	// Admins are allowed to bypass this check
	if commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil
	}

	session := commondomain.ParseSession(ctx)
	userID := uuid.MustParse(session.Subject)

	traits, err := s.kratos.FetchIdentity(ctx, userID)
	if err != nil {
		return fmt.Errorf("could not check permission for contest creation: %w", err)
	}

	oneMonthAgo := s.clock.Now().AddDate(0, -1, 0)

	if traits.CreatedAt.After(oneMonthAgo) {
		return fmt.Errorf("account too young")
	}

	contestCount, err := s.repo.GetContestsByUserCountForYear(ctx, s.clock.Now(), userID)
	if err != nil {
		return fmt.Errorf("could not check permission for contest creation: %w", err)
	}

	if contestCount >= UserCreateContestYearlyLimit {
		return fmt.Errorf("hit limit of created contests: %w", ErrForbidden)
	}

	return nil
}
