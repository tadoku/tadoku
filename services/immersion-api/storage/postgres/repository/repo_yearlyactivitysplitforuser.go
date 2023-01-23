package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) YearlyActivitySplitForUser(ctx context.Context, req *query.YearlyActivitySplitForUserRequest) (*query.YearlyActivitySplitForUserResponse, error) {
	rows, err := r.q.YearlyActivitySplitForUser(ctx, postgres.YearlyActivitySplitForUserParams{
		UserID: req.UserID,
		Year:   int16(req.Year),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.YearlyActivitySplitForUserResponse{
				Activities: []query.ActivityScore{},
			}, nil
		}
		return nil, fmt.Errorf("could not fetch activity split: %w", err)
	}

	scores := make([]query.ActivityScore, len(rows))
	for i, row := range rows {
		row := row
		scores[i] = query.ActivityScore{
			ActivityID:   int(row.LogActivityID),
			ActivityName: row.LogActivityName,
			Score:        row.Score,
		}
	}

	return &query.YearlyActivitySplitForUserResponse{
		Activities: scores,
	}, nil
}
