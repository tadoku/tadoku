package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type ContestModerationDetachLogRepository interface {
	FindContestByID(context.Context, *ContestFindRequest) (*ContestView, error)
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	DetachLogFromContest(context.Context, *ContestModerationDetachLogRequest, uuid.UUID) error
}

type ContestModerationDetachLogRequest struct {
	ContestID uuid.UUID
	LogID     uuid.UUID
	Reason    string
}

type ContestModerationDetachLog struct {
	repo ContestModerationDetachLogRepository
}

func NewContestModerationDetachLog(repo ContestModerationDetachLogRepository) *ContestModerationDetachLog {
	return &ContestModerationDetachLog{repo: repo}
}

func (s *ContestModerationDetachLog) Execute(ctx context.Context, req *ContestModerationDetachLogRequest) error {
	// Check if user is authenticated
	if commondomain.IsRole(ctx, commondomain.RoleGuest) {
		return ErrUnauthorized
	}

	if commondomain.IsRole(ctx, commondomain.RoleBanned) {
		return ErrForbidden
	}

	// Get session to extract user ID
	session := commondomain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	// Verify contest exists
	contest, err := s.repo.FindContestByID(ctx, &ContestFindRequest{
		ID:             req.ContestID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find contest: %w", err)
	}

	// Check authorization: user must be contest owner OR have Admin role
	isContestOwner := contest.OwnerUserID == userID
	isAdmin := commondomain.IsRole(ctx, commondomain.RoleAdmin)

	if !isContestOwner && !isAdmin {
		return ErrForbidden
	}

	// Verify log exists
	_, err = s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find log: %w", err)
	}

	// Detach log from contest with audit logging
	return s.repo.DetachLogFromContest(ctx, req, userID)
}
