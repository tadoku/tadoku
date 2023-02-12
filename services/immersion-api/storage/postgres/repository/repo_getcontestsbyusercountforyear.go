package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) GetContestsByUserCountForYear(ctx context.Context, now time.Time, userID uuid.UUID) (int32, error) {
	res, err := r.q.GetContestsByUserCountForYear(ctx, postgres.GetContestsByUserCountForYearParams{
		UserID: userID,
		Year:   int32(now.Year()),
	})
	if err != nil {
		return 0, fmt.Errorf("could not fetch contest contest: %w", err)
	}

	return int32(res), nil
}
