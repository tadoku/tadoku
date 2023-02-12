package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type CreateContestPermissionCheckRequest struct {
	UserID    uuid.UUID
}

var contestLimit int32 = 12

func (s *ServiceImpl) CreateContestPermissionCheck(ctx context.Context, req *CreateContestPermissionCheckRequest) (error) {
	// Admins are allowed to bypass this check
	if domain.IsRole(ctx, domain.RoleAdmin) {
		return nil
	}

	contestCount, err := s.r.GetContestsByUserCountForYear(ctx, s.clock.Now(), req.UserID)
	if err != nil {
		return fmt.Errorf("could not check permission for contest creation: %w", err)
	}

	if contestCount > contestLimit {
		return fmt.Errorf("hit limit of created contests")
	}

	return nil
}
