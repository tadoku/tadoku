package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyScoresForUser(ctx context.Context, req *domain.ProfileYearlyScoresRequest) ([]domain.Score, error) {
	rows, err := r.q.FetchScoresForProfile(ctx, postgres.FetchScoresForProfileParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.Score{}, nil
		}
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	scores := make([]domain.Score, len(rows))
	for i, row := range rows {
		scores[i] = domain.Score{
			LanguageCode: row.LanguageCode,
			LanguageName: &row.LanguageName,
			Score:        row.Score,
		}
	}

	return scores, nil
}
