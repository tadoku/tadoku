package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func findRunningOwnedContestAvailableAfter(ctx context.Context, q *postgres.Queries, userID uuid.UUID, checkedAt time.Time) (*time.Time, error) {
	availableAfter, err := q.FindRunningOwnedContestAvailableAfter(ctx, postgres.FindRunningOwnedContestAvailableAfterParams{
		UserID:    userID,
		CheckedAt: checkedAt,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	availableAfter = availableAfter.UTC()
	return &availableAfter, nil
}

func (r *Repository) FindRunningOwnedContestAvailableAfter(ctx context.Context, userID uuid.UUID, checkedAt time.Time) (*time.Time, error) {
	availableAfter, err := findRunningOwnedContestAvailableAfter(ctx, r.q, userID, checkedAt)
	if err != nil {
		return nil, fmt.Errorf("could not find running owned contest: %w", err)
	}
	return availableAfter, nil
}
