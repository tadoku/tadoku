package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogDeleteRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	DeleteLog(context.Context, *LogDeleteRequest) error
}

type LogDeleteRequest struct {
	LogID uuid.UUID
	Now   time.Time
}

type LogDelete struct {
	repo               LogDeleteRepository
	clock              commondomain.Clock
	leaderboardUpdater *LeaderboardUpdater
}

func NewLogDelete(
	repo LogDeleteRepository,
	clock commondomain.Clock,
	leaderboardUpdater *LeaderboardUpdater,
) *LogDelete {
	return &LogDelete{
		repo:               repo,
		clock:              clock,
		leaderboardUpdater: leaderboardUpdater,
	}
}

func (s *LogDelete) Execute(ctx context.Context, req *LogDeleteRequest) error {
	if err := requireAuthentication(ctx); err != nil {
		return err
	}

	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return ErrUnauthorized
	}

	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find log to delete: %w", err)
	}

	isOwner := log.UserID == uuid.MustParse(session.Subject)
	if !isOwner && !isAdmin(ctx) {
		return ErrForbidden
	}

	req.Now = s.clock.Now()

	if err := s.repo.DeleteLog(ctx, req); err != nil {
		return err
	}

	// Rebuild affected leaderboards — best effort, do not fail the deletion
	s.updateLeaderboardsAfterDelete(ctx, log)

	return nil
}

// updateLeaderboardsAfterDelete updates all leaderboards affected by a deleted log.
// For each contest the log was attached to: update that user's contest score.
// If the log was eligible for official leaderboard: update yearly and global scores.
func (s *LogDelete) updateLeaderboardsAfterDelete(ctx context.Context, log *Log) {
	for _, reg := range log.Registrations {
		s.leaderboardUpdater.UpdateUserContestScore(ctx, reg.ContestID, log.UserID)
	}

	if log.EligibleOfficialLeaderboard {
		year := log.CreatedAt.Year()
		s.leaderboardUpdater.UpdateUserOfficialScores(ctx, year, log.UserID)
	}
}
