package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type FindRegistrationForUserRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

func (s *ServiceImpl) FindRegistrationForUser(ctx context.Context, req *FindRegistrationForUserRequest) (*ContestRegistration, error) {
	session := domain.ParseSession(ctx)

	if domain.IsRole(ctx, domain.RoleGuest) || domain.IsRole(ctx, domain.RoleBanned) || session == nil {
		return nil, ErrUnauthorized
	}

	req.UserID = uuid.MustParse(session.Subject)

	res, err := s.r.FindRegistrationForUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
