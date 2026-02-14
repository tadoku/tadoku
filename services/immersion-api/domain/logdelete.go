package domain

import (
	"context"
	"fmt"
	"log/slog"
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
	repo             LogDeleteRepository
	clock            commondomain.Clock
	leaderboardStore LeaderboardStore
	leaderboardRepo  LeaderboardRebuildRepository
}

func NewLogDelete(
	repo LogDeleteRepository,
	clock commondomain.Clock,
	leaderboardStore LeaderboardStore,
	leaderboardRepo LeaderboardRebuildRepository,
) *LogDelete {
	return &LogDelete{
		repo:             repo,
		clock:            clock,
		leaderboardStore: leaderboardStore,
		leaderboardRepo:  leaderboardRepo,
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
	s.rebuildLeaderboardsAfterDelete(ctx, log)

	return nil
}

// rebuildLeaderboardsAfterDelete rebuilds all leaderboards affected by a deleted log.
// For each contest the log was attached to: rebuild that contest's leaderboard.
// If the log was eligible for official leaderboard: rebuild yearly and global.
func (s *LogDelete) rebuildLeaderboardsAfterDelete(ctx context.Context, log *Log) {
	for _, reg := range log.Registrations {
		scores, err := s.leaderboardRepo.FetchAllContestLeaderboardScores(ctx, reg.ContestID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch contest leaderboard scores for rebuild after delete", "contest_id", reg.ContestID, "error", err)
			continue
		}
		if err := s.leaderboardStore.RebuildContestLeaderboard(ctx, reg.ContestID, scores); err != nil {
			slog.ErrorContext(ctx, "failed to rebuild contest leaderboard after delete", "contest_id", reg.ContestID, "error", err)
		}
	}

	if log.EligibleOfficialLeaderboard {
		year := log.CreatedAt.Year()

		scores, err := s.leaderboardRepo.FetchAllYearlyLeaderboardScores(ctx, year)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch yearly leaderboard scores for rebuild after delete", "year", year, "error", err)
		} else if err := s.leaderboardStore.RebuildYearlyLeaderboard(ctx, year, scores); err != nil {
			slog.ErrorContext(ctx, "failed to rebuild yearly leaderboard after delete", "year", year, "error", err)
		}

		globalScores, err := s.leaderboardRepo.FetchAllGlobalLeaderboardScores(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to fetch global leaderboard scores for rebuild after delete", "error", err)
		} else if err := s.leaderboardStore.RebuildGlobalLeaderboard(ctx, globalScores); err != nil {
			slog.ErrorContext(ctx, "failed to rebuild global leaderboard after delete", "error", err)
		}
	}
}
