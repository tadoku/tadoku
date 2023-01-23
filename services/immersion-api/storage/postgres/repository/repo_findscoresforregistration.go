package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindScoresForRegistration(ctx context.Context, req *query.ContestProfileRequest) ([]query.Score, error) {
	rows, err := r.q.FetchScoresForContestProfile(ctx, postgres.FetchScoresForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]query.Score, len(rows))
	for i, row := range rows {
		scores[i] = query.Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		}
	}

	return scores, nil
}
