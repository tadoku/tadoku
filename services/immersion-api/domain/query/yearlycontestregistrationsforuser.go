package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type YearlyContestRegistrationsForUserRequest struct {
	UserID         uuid.UUID
	Year           int
	IncludePrivate bool
}

func (s *ServiceImpl) YearlyContestRegistrationsForUser(ctx context.Context, req *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error) {
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userId, err := uuid.Parse(session.Subject)

	sessionMatchesUser := err == nil && userId == req.UserID
	req.IncludePrivate = domain.IsRole(ctx, domain.RoleAdmin) || sessionMatchesUser

	return s.r.YearlyContestRegistrationsForUser(ctx, req)
}
