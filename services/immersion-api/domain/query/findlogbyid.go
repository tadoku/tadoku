package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type FindLogByIDRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

func (s *ServiceImpl) FindLogByID(ctx context.Context, req *FindLogByIDRequest) (*Log, error) {
	req.IncludeDeleted = domain.IsRole(ctx, domain.RoleAdmin)

	log, err := s.r.FindLogByID(ctx, req)
	if err != nil {
		return nil, err
	}

	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(session.Subject)

	// Needed to prevent leaking private registrations, only show to admins and the owner of the log
	if err != nil || !domain.IsRole(ctx, domain.RoleAdmin) && log.UserID != userID {
		log.Registrations = nil
	}

	return log, nil
}
