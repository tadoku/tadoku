package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *LogsRepository) UserContestLogViews(ctx context.Context, userID uuid.UUID) ([]contestLogView, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ReadUserContestLogViews(ctx, postgres.UUID(userID))
	if err != nil {
		return nil, err
	}

	views := make([]contestLogView, 0, len(rows))
	for _, row := range rows {
		views = append(views, contestLogViewFromRow(queries.ReadContestLogViewsRow(row)))
	}
	return views, nil
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
	openIfEndOnOrAfter time.Time,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	return queries.New(executor).DeleteUserContestLogsOutsideFinishedContests(
		ctx,
		queries.DeleteUserContestLogsOutsideFinishedContestsParams{
			UserID:             postgres.UUID(userID),
			OpenIfEndOnOrAfter: pgtype.Date{Time: openIfEndOnOrAfter, Valid: true},
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
