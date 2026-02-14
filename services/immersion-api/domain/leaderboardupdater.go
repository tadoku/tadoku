package domain

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// LeaderboardScoreUpdater is the interface used by domain services to update
// leaderboard scores. All operations are best-effort — errors are logged but
// never returned, since leaderboard updates should not fail the primary operation.
type LeaderboardScoreUpdater interface {
	UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID)
	UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID)
}

// LeaderboardUpdater implements LeaderboardScoreUpdater and also provides
// full rebuild methods for reconciliation.
type LeaderboardUpdater struct {
	store LeaderboardStore
	repo  LeaderboardRepository
}

func NewLeaderboardUpdater(store LeaderboardStore, repo LeaderboardRepository) *LeaderboardUpdater {
	return &LeaderboardUpdater{
		store: store,
		repo:  repo,
	}
}

// UpdateUserContestScore recalculates a user's total score for a contest
// from the database and updates it in the store. If the leaderboard doesn't
// exist in the store yet, a full rebuild is performed.
func (u *LeaderboardUpdater) UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) {
	score, err := u.repo.FetchUserContestScore(ctx, contestID, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch user contest score", "contest_id", contestID, "user_id", userID, "error", err)
		return
	}

	updated, err := u.store.UpdateContestScore(ctx, contestID, userID, score)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update user contest score", "contest_id", contestID, "user_id", userID, "error", err)
		return
	}
	if updated {
		return
	}

	// Leaderboard doesn't exist in store yet, do a full rebuild
	u.RebuildContestLeaderboard(ctx, contestID)
}

// UpdateUserOfficialScores recalculates a user's yearly and global scores
// from the database and updates them in the store atomically. If either
// leaderboard doesn't exist, a full rebuild of that leaderboard is performed.
func (u *LeaderboardUpdater) UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID) {
	yearlyScore, err := u.repo.FetchUserYearlyScore(ctx, year, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch user yearly score", "year", year, "user_id", userID, "error", err)
		return
	}

	globalScore, err := u.repo.FetchUserGlobalScore(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch user global score", "user_id", userID, "error", err)
		return
	}

	yearlyUpdated, globalUpdated, err := u.store.UpdateOfficialScores(ctx, year, userID, yearlyScore, globalScore)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update user official scores", "year", year, "user_id", userID, "error", err)
		return
	}

	// If either leaderboard didn't exist, rebuild both together
	if !yearlyUpdated || !globalUpdated {
		u.RebuildOfficialLeaderboards(ctx, year)
	}
}

// RebuildContestLeaderboard fetches all scores from the database and
// fully rebuilds the contest leaderboard in the store.
func (u *LeaderboardUpdater) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID) {
	scores, err := u.repo.FetchAllContestLeaderboardScores(ctx, contestID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch contest leaderboard scores for rebuild", "contest_id", contestID, "error", err)
		return
	}

	if err := u.store.RebuildContestLeaderboard(ctx, contestID, scores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild contest leaderboard", "contest_id", contestID, "error", err)
	}
}

// RebuildOfficialLeaderboards fetches all scores from the database and
// atomically rebuilds both yearly and global leaderboards in the store.
func (u *LeaderboardUpdater) RebuildOfficialLeaderboards(ctx context.Context, year int) {
	yearlyScores, err := u.repo.FetchAllYearlyLeaderboardScores(ctx, year)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch yearly leaderboard scores for rebuild", "year", year, "error", err)
		return
	}

	globalScores, err := u.repo.FetchAllGlobalLeaderboardScores(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch global leaderboard scores for rebuild", "error", err)
		return
	}

	if err := u.store.RebuildOfficialLeaderboards(ctx, year, yearlyScores, globalScores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild official leaderboards", "year", year, "error", err)
	}
}
