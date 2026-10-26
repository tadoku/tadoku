package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) UserContestLogRemovals(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]contestevent.Change, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ListUserContestLogRemovals(ctx, queries.ListUserContestLogRemovalsParams{
		UserID: postgres.UUID(userID),
		Now:    pgtype.Date{Time: now, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	changes := make([]contestevent.Change, 0, len(rows))
	for _, row := range rows {
		change := contestevent.Change{
			ContestID: uuid.UUID(row.ContestID.Bytes),
			LogID:     uuid.UUID(row.LogID.Bytes),
			Type:      contestevent.Updated,
		}
		if row.Withdrawn {
			change.Type = contestevent.Withdrawn
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (r *LogsRepository) UserOfficialYears(ctx context.Context, userID uuid.UUID) ([]int16, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return queries.New(executor).ListOfficialYearsForUser(ctx, postgres.UUID(userID))
}

func (r *LogsRepository) DeleteUserContestLogsOutsideFinishedContests(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	return queries.New(executor).DeleteUserContestLogsOutsideFinishedContests(
		ctx,
		queries.DeleteUserContestLogsOutsideFinishedContestsParams{
			UserID: postgres.UUID(userID),
			Now:    pgtype.Date{Time: now, Valid: true},
		},
	)
}

func (r *LogsRepository) FreezeUserContestHistory(ctx context.Context, userID uuid.UUID, now time.Time) (int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	frozen, err := queries.New(executor).FreezeUserContestHistory(ctx, queries.FreezeUserContestHistoryParams{
		FrozenAt: postgres.Timestamp(now),
		UserID:   postgres.UUID(userID),
	})
	return int(frozen), err
}

func (r *LogsRepository) DeleteUserTags(ctx context.Context, userID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	return queries.New(executor).DeleteUserLogTags(ctx, postgres.UUID(userID))
}

func (r *LogsRepository) DeleteUserLogsOutsideContests(ctx context.Context, userID uuid.UUID) (int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}
	deleted, err := queries.New(executor).DeleteUserLogsOutsideContests(ctx, postgres.UUID(userID))
	return int(deleted), err
}
