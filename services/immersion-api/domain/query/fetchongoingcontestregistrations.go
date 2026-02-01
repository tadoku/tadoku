package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type FetchOngoingContestRegistrationsRequest struct {
	UserID uuid.UUID
	Now    time.Time
}

func (s *ServiceImpl) FetchOngoingContestRegistrations(ctx context.Context, req *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	session := domain.ParseSession(ctx)
	req.UserID = uuid.MustParse(session.Subject)
	req.Now = s.clock.Now()

	return s.r.FetchOngoingContestRegistrations(ctx, req)
}
