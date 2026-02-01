package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
)

type DeleteLogRequest struct {
	LogID uuid.UUID
	Now   time.Time
}

func (s *ServiceImpl) DeleteLog(ctx context.Context, req *DeleteLogRequest) error {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return ErrUnauthorized
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}

	log, err := s.r.FindLogByID(ctx, &immersiondomain.LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find log to deletel: %w", err)
	}

	isOwner := log.UserID == uuid.MustParse(session.Subject)
	if !isOwner && !domain.IsRole(ctx, domain.RoleAdmin) {
		return ErrForbidden
	}

	req.Now = s.clock.Now()

	return s.r.DeleteLog(ctx, req)
}
