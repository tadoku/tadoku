package contests

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *ContestsRepository) DeleteUnfinishedRegistrations(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]uuid.UUID, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).DeleteUnfinishedRegistrationsForUser(
		ctx,
		queries.DeleteUnfinishedRegistrationsForUserParams{
			UserID: postgres.UUID(userID),
			Now:    pgtype.Date{Time: now, Valid: true},
		},
	)
	if err != nil {
		return nil, err
	}
	return postgres.UUIDs(rows), nil
}

func (r *ContestsRepository) CountRegistrations(ctx context.Context, userID uuid.UUID) (int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	count, err := queries.New(executor).CountRegistrationsForUser(ctx, postgres.UUID(userID))
	return int(count), err
}

func (r *ContestsRepository) RenameContestsOwnedBy(ctx context.Context, userID uuid.UUID) (int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	renamed, err := queries.New(executor).RenameContestsOwnedByUser(ctx, postgres.UUID(userID))
	return int(renamed), err
}
