package logs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type LogsRepository struct{ db *pgxpool.Pool }

func NewLogsRepository(db *pgxpool.Pool) *LogsRepository { return &LogsRepository{db: db} }

func (r *LogsRepository) LockLog(ctx context.Context, id uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	frozen, err := queries.New(executor).LockLogForMutation(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLogNotFound
	}
	if err != nil {
		return fmt.Errorf("lock log: %w", err)
	}
	if frozen.Valid {
		return ErrLogFrozen
	}
	return nil
}

func (r *LogsRepository) CreateLog(ctx context.Context, mutation logMutation) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(mutation.Tracking)
	return queries.New(executor).CreateLog(ctx, queries.CreateLogParams{
		ID:                          postgres.UUID(mutation.ID),
		UserID:                      postgres.UUID(mutation.UserID),
		LanguageCode:                mutation.LanguageCode,
		ActivityID:                  int16(mutation.ActivityID),
		UnitID:                      p.unitID,
		UnitKey:                     p.unitKey,
		Amount:                      p.amount,
		Modifier:                    p.modifier,
		DurationSeconds:             p.duration,
		ComputedScore:               pgtype.Float4{Float32: mutation.Tracking.Score, Valid: true},
		ScoreRuleSetID:              p.ruleSetID,
		ScoreRuleIds:                p.ruleIDs,
		ScoreRates:                  mutation.Tracking.Rates,
		ScoreSource:                 p.source,
		EligibleOfficialLeaderboard: mutation.EligibleOfficialLeaderboard,
		Description:                 postgres.NullableText(mutation.Description),
		CreatedAt:                   postgres.Timestamp(mutation.Now),
		UpdatedAt:                   postgres.Timestamp(mutation.Now),
	})
}

func (r *LogsRepository) CreateContestLog(ctx context.Context, logID uuid.UUID, tracking logscore.ContestTracking) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(tracking.Tracking)
	return queries.New(executor).CreateContestLog(ctx, queries.CreateContestLogParams{
		RegistrationID:  postgres.UUID(tracking.RegistrationID),
		LogID:           postgres.UUID(logID),
		UnitKey:         p.unitKey,
		Amount:          p.amount,
		Modifier:        p.modifier,
		DurationSeconds: p.duration,
		ComputedScore:   pgtype.Float4{Float32: tracking.Tracking.Score, Valid: true},
		ScoreRuleSetID:  p.ruleSetID,
		ScoreRuleIds:    p.ruleIDs,
		ScoreRates:      tracking.Tracking.Rates,
		ScoreSource:     p.source,
	})
}

func (r *LogsRepository) InsertTag(ctx context.Context, logID, userID uuid.UUID, tag string) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	return queries.New(executor).InsertLogTag(ctx, queries.InsertLogTagParams{
		LogID:  postgres.UUID(logID),
		UserID: postgres.UUID(userID),
		Tag:    tag,
	})
}
func (r *LogsRepository) DeleteTags(ctx context.Context, logID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	return queries.New(executor).DeleteLogTags(ctx, postgres.UUID(logID))
}

func (r *LogsRepository) JobContext(ctx context.Context, id uuid.UUID) (JobContext, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return JobContext{}, err
	}
	row, err := queries.New(executor).FetchLogJobContext(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return JobContext{}, ErrLogNotFound
	}
	if err != nil {
		return JobContext{}, err
	}
	return JobContext{
		UserID:           row.UserID.Bytes,
		Year:             row.Year,
		EligibleOfficial: row.EligibleOfficialLeaderboard,
	}, nil
}

func (r *LogsRepository) UpdateLog(ctx context.Context, mutation logMutation) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(mutation.Tracking)
	return queries.New(executor).UpdateLog(ctx, queries.UpdateLogParams{
		UnitID:          p.unitID,
		UnitKey:         p.unitKey,
		Amount:          p.amount,
		Modifier:        p.modifier,
		DurationSeconds: p.duration,
		ComputedScore:   pgtype.Float4{Float32: mutation.Tracking.Score, Valid: true},
		ScoreRuleSetID:  p.ruleSetID,
		ScoreRuleIds:    p.ruleIDs,
		ScoreRates:      mutation.Tracking.Rates,
		ScoreSource:     p.source,
		Description:     postgres.NullableText(mutation.Description),
		UpdatedAt:       postgres.Timestamp(mutation.Now),
		LogID:           postgres.UUID(mutation.ID),
	})
}
func (r *LogsRepository) UpdateContestLog(
	ctx context.Context,
	logID uuid.UUID,
	tracking logscore.ContestTracking,
	openIfEndOnOrAfter time.Time,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(tracking.Tracking)
	return queries.New(executor).UpdateOngoingContestLog(ctx, queries.UpdateOngoingContestLogParams{
		UnitKey:            p.unitKey,
		Amount:             p.amount,
		Modifier:           p.modifier,
		DurationSeconds:    p.duration,
		ComputedScore:      pgtype.Float4{Float32: tracking.Tracking.Score, Valid: true},
		ScoreRuleSetID:     p.ruleSetID,
		ScoreRuleIds:       p.ruleIDs,
		ScoreRates:         tracking.Tracking.Rates,
		ScoreSource:        p.source,
		LogID:              postgres.UUID(logID),
		ContestID:          postgres.UUID(tracking.ContestID),
		OpenIfEndOnOrAfter: pgtype.Date{Time: openIfEndOnOrAfter, Valid: true},
	})
}

func (r *LogsRepository) UpdateOngoingContestLogs(
	ctx context.Context,
	logID uuid.UUID,
	tracking logscore.Tracking,
	openIfEndOnOrAfter time.Time,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(tracking)
	return queries.New(executor).UpdateOngoingContestLogs(ctx, queries.UpdateOngoingContestLogsParams{
		UnitKey:            p.unitKey,
		Amount:             p.amount,
		Modifier:           p.modifier,
		DurationSeconds:    p.duration,
		ComputedScore:      pgtype.Float4{Float32: tracking.Score, Valid: true},
		ScoreRuleSetID:     p.ruleSetID,
		ScoreRuleIds:       p.ruleIDs,
		ScoreRates:         tracking.Rates,
		ScoreSource:        p.source,
		LogID:              postgres.UUID(logID),
		OpenIfEndOnOrAfter: pgtype.Date{Time: openIfEndOnOrAfter, Valid: true},
	})
}

type logTrackingParams struct {
	unitID    pgtype.UUID
	unitKey   pgtype.Text
	amount    pgtype.Float4
	modifier  pgtype.Float4
	duration  pgtype.Int4
	ruleSetID pgtype.UUID
	ruleIDs   []pgtype.UUID
	source    pgtype.Text
}

func trackingParams(t logscore.Tracking) logTrackingParams {
	p := logTrackingParams{
		unitKey:   postgres.NullableNonEmptyText(&t.UnitKey),
		amount:    postgres.NullableFloat4(t.Amount),
		modifier:  postgres.NullableFloat4(t.Modifier),
		duration:  postgres.NullableInt4(t.DurationSeconds),
		source:    postgres.NullableNonEmptyText(&t.Source),
		unitID:    postgres.NullableUUID(t.UnitID),
		ruleSetID: postgres.NullableUUID(t.RuleSetID),
	}
	for _, id := range t.RuleIDs {
		p.ruleIDs = append(p.ruleIDs, postgres.UUID(id))
	}
	return p
}

func (r *LogsRepository) DetachContest(ctx context.Context, logID, contestID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	return queries.New(executor).DetachContestLog(ctx, queries.DetachContestLogParams{
		LogID:     postgres.UUID(logID),
		ContestID: postgres.UUID(contestID),
	})
}

func (r *LogsRepository) RecomputeOfficialEligibility(ctx context.Context, logID uuid.UUID, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	return queries.New(executor).RecomputeLogOfficialEligibility(ctx, queries.RecomputeLogOfficialEligibilityParams{
		LogID:     postgres.UUID(logID),
		UpdatedAt: postgres.Timestamp(now),
	})
}

func (r *LogsRepository) AttachedContests(ctx context.Context, logID uuid.UUID) ([]attachedContest, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListAttachedContests(ctx, postgres.UUID(logID))
	if err != nil {
		return nil, err
	}

	contests := make([]attachedContest, 0, len(rows))
	for _, row := range rows {
		contests = append(contests, attachedContest{ID: row.ContestID.Bytes, End: row.ContestEnd.Time})
	}
	return contests, nil
}

func (r *LogsRepository) SoftDelete(ctx context.Context, logID uuid.UUID, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	return queries.New(executor).SoftDeleteLog(ctx, queries.SoftDeleteLogParams{
		LogID:     postgres.UUID(logID),
		DeletedAt: postgres.Timestamp(now),
	})
}

func (r *LogsRepository) AttachedRegistrations(ctx context.Context, id uuid.UUID) ([]RegistrationReference, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).FindAttachedContestRegistrationsForLog(ctx, postgres.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("find attached registrations: %w", err)
	}

	result := make([]RegistrationReference, 0, len(rows))
	for _, row := range rows {
		result = append(result, RegistrationReference{
			RegistrationID:       uuid.UUID(row.ID.Bytes),
			ContestID:            uuid.UUID(row.ContestID.Bytes),
			ContestEnd:           row.ContestEnd.Time,
			Title:                row.Title,
			OwnerUserDisplayName: row.OwnerUserDisplayName,
			Official:             row.Official,
			Score:                row.Score.Float32,
		})
	}
	return result, nil
}
