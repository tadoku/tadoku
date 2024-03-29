package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var UserCreateContestYearlyLimit int32 = 12

func (s *ServiceImpl) CreateContestPermissionCheck(ctx context.Context) error {
	// Admins are allowed to bypass this check
	if domain.IsRole(ctx, domain.RoleAdmin) {
		return nil
	}

	session := domain.ParseSession(ctx)
	userID := uuid.MustParse(session.Subject)

	traits, err := s.kratos.FetchIdentity(ctx, userID)
	if err != nil {
		return fmt.Errorf("could not check permission for contest creation: %w", err)
	}

	oneMonthAgo := s.clock.Now().AddDate(0, -1, 0)

	if traits.CreatedAt.After(oneMonthAgo) {
		return fmt.Errorf("account too young")
	}

	contestCount, err := s.r.GetContestsByUserCountForYear(ctx, s.clock.Now(), userID)
	if err != nil {
		return fmt.Errorf("could not check permission for contest creation: %w", err)
	}

	if contestCount >= UserCreateContestYearlyLimit {
		return fmt.Errorf("hit limit of created contests: %w", ErrForbidden)
	}

	return nil
}
