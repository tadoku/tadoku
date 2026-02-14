package domain

import (
	"context"
	"fmt"
	"log/slog"

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
	repo             ContestModerationDetachLogRepository
	leaderboardStore LeaderboardStore
	leaderboardRepo  LeaderboardRebuildRepository
}

func NewContestModerationDetachLog(
	repo ContestModerationDetachLogRepository,
	leaderboardStore LeaderboardStore,
	leaderboardRepo LeaderboardRebuildRepository,
) *ContestModerationDetachLog {
	return &ContestModerationDetachLog{
		repo:             repo,
		leaderboardStore: leaderboardStore,
		leaderboardRepo:  leaderboardRepo,
	}
}

func (s *ContestModerationDetachLog) Execute(ctx context.Context, req *ContestModerationDetachLogRequest) error {
	if err := requireAuthentication(ctx); err != nil {
		return err
	}

	// Get session to extract user ID
	session := commondomain.ParseUserIdentity(ctx)
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
	isAdmin := isAdmin(ctx)

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
	if err := s.repo.DetachLogFromContest(ctx, req, userID); err != nil {
		return err
	}

	// Rebuild contest leaderboard — best effort, do not fail the detach
	s.rebuildContestLeaderboard(ctx, req.ContestID)

	return nil
}

func (s *ContestModerationDetachLog) rebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID) {
	scores, err := s.leaderboardRepo.FetchAllContestLeaderboardScores(ctx, contestID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch contest leaderboard scores for rebuild after detach", "contest_id", contestID, "error", err)
		return
	}
	if err := s.leaderboardStore.RebuildContestLeaderboard(ctx, contestID, scores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild contest leaderboard after detach", "contest_id", contestID, "error", err)
	}
}
