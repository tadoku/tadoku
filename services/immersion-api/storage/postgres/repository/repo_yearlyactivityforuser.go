package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyActivityForUser(ctx context.Context, req *domain.ProfileYearlyActivityRequest) ([]domain.UserActivityScore, error) {
	rows, err := r.q.YearlyActivityForUser(ctx, postgres.YearlyActivityForUserParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.UserActivityScore{}, nil
		}
		return nil, fmt.Errorf("could not fetch activity summary: %w", err)
	}

	res := make([]domain.UserActivityScore, len(rows))
	for i, it := range rows {
		res[i] = domain.UserActivityScore{
			Date:    it.Date,
			Score:   it.Score,
			Updates: int(it.UpdateCount),
		}
	}

	return res, nil
}
