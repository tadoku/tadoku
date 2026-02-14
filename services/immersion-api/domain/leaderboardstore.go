package domain

import (
	"context"

	"github.com/google/uuid"
)

// LeaderboardScore represents a user's total score in a leaderboard.
type LeaderboardScore struct {
	UserID uuid.UUID
	Score  float64
}

// LeaderboardStore manages leaderboard sorted sets in a fast key-value store.
// Only unfiltered (main) leaderboards are stored — no language or activity filters.
type LeaderboardStore interface {
	// UpdateContestScore sets a user's absolute score in a contest leaderboard.
	// Only updates if the leaderboard already exists in the store.
	// Returns true if the leaderboard existed and the score was set.
	UpdateContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error)

	// UpdateOfficialScores atomically sets a user's yearly and global scores.
	// Only updates leaderboards that already exist in the store.
	// Returns which leaderboards were updated (yearlyUpdated, globalUpdated).
	UpdateOfficialScores(ctx context.Context, year int, userID uuid.UUID, yearlyScore float64, globalScore float64) (yearlyUpdated bool, globalUpdated bool, err error)

	// RebuildContestLeaderboard atomically replaces a contest leaderboard with the given scores.
	RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []LeaderboardScore) error

	// RebuildOfficialLeaderboards atomically replaces both yearly and global leaderboards.
	RebuildOfficialLeaderboards(ctx context.Context, year int, yearlyScores []LeaderboardScore, globalScores []LeaderboardScore) error
}

// LeaderboardRepository provides queries for fetching leaderboard scores
// from the database, used for updating and rebuilding leaderboards in the store.
type LeaderboardRepository interface {
	FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]LeaderboardScore, error)
	FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]LeaderboardScore, error)
	FetchAllGlobalLeaderboardScores(ctx context.Context) ([]LeaderboardScore, error)
	FetchUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) (float64, error)
	FetchUserYearlyScore(ctx context.Context, year int, userID uuid.UUID) (float64, error)
	FetchUserGlobalScore(ctx context.Context, userID uuid.UUID) (float64, error)
}
