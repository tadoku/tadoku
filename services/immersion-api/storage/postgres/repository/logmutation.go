package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func lockLogForMutation(ctx context.Context, qtx *postgres.Queries, logID uuid.UUID) error {
	frozenAt, err := qtx.LockLogForMutation(ctx, logID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if frozenAt.Valid {
		return domain.ErrLogFrozen
	}
	return nil
}
