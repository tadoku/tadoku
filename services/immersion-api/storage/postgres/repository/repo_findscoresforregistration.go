package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindScoresForRegistration(ctx context.Context, req *domain.ProfileContestRequest) ([]domain.Score, error) {
	rows, err := r.q.FetchScoresForContestProfile(ctx, postgres.FetchScoresForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]domain.Score, len(rows))
	for i, row := range rows {
		scores[i] = domain.Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		}
	}

	return scores, nil
}
