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
	// IncrementContestScore atomically increments a user's score in a contest leaderboard.
	// Returns true if the leaderboard existed and the score was incremented.
	// Returns false if the leaderboard does not exist yet (needs rebuild).
	IncrementContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error)

	// IncrementYearlyScore atomically increments a user's score in a yearly leaderboard.
	// Returns true if the leaderboard existed and the score was incremented.
	// Returns false if the leaderboard does not exist yet (needs rebuild).
	IncrementYearlyScore(ctx context.Context, year int, userID uuid.UUID, score float64) (bool, error)

	// IncrementGlobalScore atomically increments a user's score in the global leaderboard.
	// Returns true if the leaderboard existed and the score was incremented.
	// Returns false if the leaderboard does not exist yet (needs rebuild).
	IncrementGlobalScore(ctx context.Context, userID uuid.UUID, score float64) (bool, error)

	// RebuildContestLeaderboard atomically replaces a contest leaderboard with the given scores.
	RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []LeaderboardScore) error

	// RebuildYearlyLeaderboard atomically replaces a yearly leaderboard with the given scores.
	RebuildYearlyLeaderboard(ctx context.Context, year int, scores []LeaderboardScore) error

	// RebuildGlobalLeaderboard atomically replaces the global leaderboard with the given scores.
	RebuildGlobalLeaderboard(ctx context.Context, scores []LeaderboardScore) error
}

// LeaderboardRebuildRepository provides queries for fetching all leaderboard scores
// from the database, used when a leaderboard needs to be rebuilt in the store.
type LeaderboardRebuildRepository interface {
	FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]LeaderboardScore, error)
	FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]LeaderboardScore, error)
	FetchAllGlobalLeaderboardScores(ctx context.Context) ([]LeaderboardScore, error)
}
