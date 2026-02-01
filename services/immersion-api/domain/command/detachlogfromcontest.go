package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
)

type DetachLogFromContestRequest struct {
	ContestID uuid.UUID
	LogID     uuid.UUID
	Reason    string
}

func (s *ServiceImpl) DetachLogFromContest(ctx context.Context, req *DetachLogFromContestRequest) error {
	// Check if user is authenticated
	if domain.IsRole(ctx, domain.RoleGuest) {
		return ErrUnauthorized
	}

	if domain.IsRole(ctx, domain.RoleBanned) {
		return ErrForbidden
	}

	// Get session to extract user ID
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	// Verify contest exists
	contest, err := s.r.FindContestByID(ctx, &immersiondomain.ContestFindRequest{
		ID:             req.ContestID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find contest: %w", err)
	}

	// Check authorization: user must be contest owner OR have Admin role
	isContestOwner := contest.OwnerUserID == userID
	isAdmin := domain.IsRole(ctx, domain.RoleAdmin)

	if !isContestOwner && !isAdmin {
		return ErrForbidden
	}

	// Verify log exists
	_, err = s.r.FindLogByID(ctx, &immersiondomain.LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find log: %w", err)
	}

	// Detach log from contest with audit logging
	return s.r.DetachLogFromContest(ctx, req, userID)
}
