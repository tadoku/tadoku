package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ActivityForContestUser(ctx context.Context, req *query.ActivityForContestUserRequest) ([]query.ActivityForContestUserRow, error) {
	rows, err := r.q.ActivityPerLanguageForContestProfile(ctx, postgres.ActivityPerLanguageForContestProfileParams{
		ContestID: req.ContestID,
		UserID:    req.UserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []query.ActivityForContestUserRow{}, nil
		}
		return nil, fmt.Errorf("could not fetch activity: %w", err)
	}

	res := make([]query.ActivityForContestUserRow, len(rows))
	for i, it := range rows {
		res[i] = query.ActivityForContestUserRow{
			Date:         it.Date,
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
		}
	}

	return res, nil
}
