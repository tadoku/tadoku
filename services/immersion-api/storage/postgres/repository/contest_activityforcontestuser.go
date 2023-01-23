package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// TODO: Rename
func (r *Repository) ReadingActivityForContestUser(ctx context.Context, req *query.ContestProfileRequest) ([]query.ReadingActivityRow, error) {
	rows, err := r.q.ReadingActivityPerLanguageForContestProfile(ctx, postgres.ReadingActivityPerLanguageForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.ReadingActivityRow{}, nil
		}
		return nil, fmt.Errorf("could not fetch reading activity: %w", err)
	}

	res := make([]query.ReadingActivityRow, len(rows))
	for i, it := range rows {
		res[i] = query.ReadingActivityRow{
			Date:         it.Date,
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
		}
	}

	return res, nil
}
