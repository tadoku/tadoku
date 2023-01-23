package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyScoresForUser(ctx context.Context, req *query.YearlyScoresForUserRequest) ([]query.Score, error) {
	rows, err := r.q.FetchScoresForProfile(ctx, postgres.FetchScoresForProfileParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.Score{}, nil
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]query.Score, len(rows))
	for i, row := range rows {
		row := row
		scores[i] = query.Score{
			LanguageCode: row.LanguageCode,
			LanguageName: &row.LanguageName,
			Score:        row.Score,
		}
	}

	return scores, nil
}
