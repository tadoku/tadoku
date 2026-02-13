package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func (r *Repository) FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]domain.LeaderboardScore, error) {
	rows, err := r.q.ContestLeaderboardAllScores(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch contest leaderboard scores: %w", err)
	}

	scores := make([]domain.LeaderboardScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.LeaderboardScore{
			UserID: row.UserID,
			Score:  float64(row.Score),
		}
	}

	return scores, nil
}

func (r *Repository) FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]domain.LeaderboardScore, error) {
	rows, err := r.q.YearlyLeaderboardAllScores(ctx, int16(year))
	if err != nil {
		return nil, fmt.Errorf("could not fetch yearly leaderboard scores: %w", err)
	}

	scores := make([]domain.LeaderboardScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.LeaderboardScore{
			UserID: row.UserID,
			Score:  float64(row.Score),
		}
	}

	return scores, nil
}

func (r *Repository) FetchAllGlobalLeaderboardScores(ctx context.Context) ([]domain.LeaderboardScore, error) {
	rows, err := r.q.GlobalLeaderboardAllScores(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch global leaderboard scores: %w", err)
	}

	scores := make([]domain.LeaderboardScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.LeaderboardScore{
			UserID: row.UserID,
			Score:  float64(row.Score),
		}
	}

	return scores, nil
}
