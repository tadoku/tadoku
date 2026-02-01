package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyActivitySplitForUser(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error) {
	rows, err := r.q.YearlyActivitySplitForUser(ctx, postgres.YearlyActivitySplitForUserParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.ProfileYearlyActivitySplitResponse{
				Activities: []domain.ActivityScore{},
			}, nil
		}
		return nil, fmt.Errorf("could not fetch activity split: %w", err)
	}

	scores := make([]domain.ActivityScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.ActivityScore{
			ActivityID:   int(row.LogActivityID),
			ActivityName: row.LogActivityName,
			Score:        row.Score,
		}
	}

	return &domain.ProfileYearlyActivitySplitResponse{
		Activities: scores,
	}, nil
}
