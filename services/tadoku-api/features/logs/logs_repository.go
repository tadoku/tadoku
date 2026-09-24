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
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/leaderboardoutbox"
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

func (r *LogsRepository) CreateContestLog(ctx context.Context, logID uuid.UUID, tracking ContestTracking) error {
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

func (r *LogsRepository) InsertOutbox(ctx context.Context, userID uuid.UUID, contestID *uuid.UUID, year *int16, event leaderboardoutbox.EventType) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	var y pgtype.Int2
	if year != nil {
		y = pgtype.Int2{Int16: *year, Valid: true}
	}
	return queries.New(executor).InsertLogLeaderboardOutbox(ctx, queries.InsertLogLeaderboardOutboxParams{
		EventType: string(event),
		UserID:    postgres.UUID(userID),
		ContestID: postgres.NullableUUID(contestID),
		Year:      y,
	})
}

func (r *LogsRepository) OutboxContext(ctx context.Context, id uuid.UUID) (OutboxContext, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return OutboxContext{}, err
	}
	row, err := queries.New(executor).FetchLogOutboxContext(ctx, postgres.UUID(id))
	return OutboxContext{
		UserID:           row.UserID.Bytes,
		Year:             row.Year,
		EligibleOfficial: row.EligibleOfficialLeaderboard,
	}, err
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
func (r *LogsRepository) UpdateContestLog(ctx context.Context, logID uuid.UUID, tracking ContestTracking, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(tracking.Tracking)
	return queries.New(executor).UpdateOngoingContestLog(ctx, queries.UpdateOngoingContestLogParams{
		UnitKey:         p.unitKey,
		Amount:          p.amount,
		Modifier:        p.modifier,
		DurationSeconds: p.duration,
		ComputedScore:   pgtype.Float4{Float32: tracking.Tracking.Score, Valid: true},
		ScoreRuleSetID:  p.ruleSetID,
		ScoreRuleIds:    p.ruleIDs,
		ScoreRates:      tracking.Tracking.Rates,
		ScoreSource:     p.source,
		LogID:           postgres.UUID(logID),
		ContestID:       postgres.UUID(tracking.ContestID),
		Now:             pgtype.Date{Time: now, Valid: true},
	})
}

func (r *LogsRepository) UpdateOngoingContestLogs(ctx context.Context, logID uuid.UUID, tracking Tracking, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	p := trackingParams(tracking)
	return queries.New(executor).UpdateOngoingContestLogs(ctx, queries.UpdateOngoingContestLogsParams{
		UnitKey:         p.unitKey,
		Amount:          p.amount,
		Modifier:        p.modifier,
		DurationSeconds: p.duration,
		ComputedScore:   pgtype.Float4{Float32: tracking.Score, Valid: true},
		ScoreRuleSetID:  p.ruleSetID,
		ScoreRuleIds:    p.ruleIDs,
		ScoreRates:      tracking.Rates,
		ScoreSource:     p.source,
		LogID:           postgres.UUID(logID),
		Now:             pgtype.Date{Time: now, Valid: true},
	})
}

func (r *LogsRepository) OngoingContestIDs(ctx context.Context, id uuid.UUID, now time.Time) ([]uuid.UUID, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).FetchOngoingContestIDsForLog(ctx, queries.FetchOngoingContestIDsForLogParams{
		LogID: postgres.UUID(id),
		Now:   pgtype.Date{Time: now, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		result[i] = row.Bytes
	}
	return result, nil
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

func trackingParams(t Tracking) logTrackingParams {
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

func (r *LogsRepository) ListUnits(ctx context.Context) ([]Unit, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list log units: %w", err)
	}

	result := make([]Unit, 0, len(rows))
	for _, row := range rows {
		var languageCode *string
		if row.LanguageCode.Valid {
			languageCode = &row.LanguageCode.String
		}
		result = append(result, Unit{
			ID:            uuid.UUID(row.ID.Bytes),
			Key:           row.UnitKey,
			LogActivityID: int(row.LogActivityID),
			Name:          row.Name,
			Modifier:      row.Modifier,
			LanguageCode:  languageCode,
		})
	}
	return result, nil
}

func (r *LogsRepository) ListUserLanguageCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	codes, err := queries.New(executor).ListDistinctLanguageCodesForUser(ctx, postgres.UUID(userID))
	if err != nil {
		return nil, fmt.Errorf("list user log languages: %w", err)
	}
	return codes, nil
}

func (r *LogsRepository) ListTagSuggestions(ctx context.Context, userID uuid.UUID, query string) ([]TagSuggestion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListTagSuggestionsForUser(ctx, queries.ListTagSuggestionsForUserParams{
		UserID: postgres.UUID(userID),
		Query:  pgtype.Text{String: query, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list tag suggestions: %w", err)
	}

	result := make([]TagSuggestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, TagSuggestion{Tag: row.Tag, Count: int(row.UsageCount)})
	}
	return result, nil
}

func (r *LogsRepository) YearlyActivity(ctx context.Context, userID uuid.UUID, year int16) ([]ActivityScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivityForUser(ctx, queries.YearlyActivityForUserParams{
		UserID: postgres.UUID(userID),
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyActivity: %w", err)
	}

	result := make([]ActivityScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ActivityScore{
			Date:    row.Date.Time,
			Score:   row.Score,
			Updates: int(row.UpdateCount),
		})
	}

	return result, nil
}

func (r *LogsRepository) YearlyScores(ctx context.Context, userID uuid.UUID, year int16) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).FetchScoresForProfile(ctx, queries.FetchScoresForProfileParams{
		UserID: postgres.UUID(userID),
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyScores: %w", err)
	}

	result := make([]Score, 0, len(rows))
	for _, row := range rows {
		result = append(result, Score{
			LanguageCode: row.LanguageCode,
			LanguageName: row.LanguageName,
			Score:        row.Score,
		})
	}

	return result, nil
}

func (r *LogsRepository) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int16) ([]ActivitySplitScore, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).YearlyActivitySplitForUser(ctx, queries.YearlyActivitySplitForUserParams{
		UserID: postgres.UUID(userID),
		Year:   year,
	})
	if err != nil {
		return nil, fmt.Errorf("YearlyActivitySplit: %w", err)
	}

	result := make([]ActivitySplitScore, 0, len(rows))
	for _, row := range rows {
		result = append(result, ActivitySplitScore{
			ActivityID: int(row.LogActivityID),
			Score:      row.Score,
		})
	}

	return result, nil
}

func (r *LogsRepository) ContestScores(ctx context.Context, userID, contestID uuid.UUID) ([]Score, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).FetchScoresForContestProfile(ctx, queries.FetchScoresForContestProfileParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if err != nil {
		return nil, fmt.Errorf("ContestScores: %w", err)
	}

	result := make([]Score, 0, len(rows))
	for _, row := range rows {
		result = append(result, Score{
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		})
	}
	return result, nil
}

func (r *LogsRepository) ContestActivity(ctx context.Context, userID, contestID uuid.UUID) ([]ContestActivity, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ActivityPerLanguageForContestProfile(ctx, queries.ActivityPerLanguageForContestProfileParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if err != nil {
		return nil, fmt.Errorf("ContestActivity: %w", err)
	}

	result := make([]ContestActivity, 0, len(rows))
	for _, row := range rows {
		result = append(result, ContestActivity{
			Date:         row.Date.Time,
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		})
	}
	return result, nil
}

func (r *LogsRepository) FindLog(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Log, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindLogByID(ctx, queries.FindLogByIDParams{
		ID:             postgres.UUID(id),
		IncludeDeleted: includeDeleted,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLogNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find log: %w", err)
	}

	tracking := Tracking{
		DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
		Score:           row.Score.Float32,
		RuleSetID:       postgres.UUIDPointer(row.ScoreRuleSetID),
		RuleIDs:         postgres.UUIDs(row.ScoreRuleIds),
		Rates:           row.ScoreRates,
	}
	if row.ScoreSource.Valid {
		tracking.Source = row.ScoreSource.String
	}
	if row.Amount.Valid && row.Modifier.Valid {
		tracking.UnitID = postgres.UUIDPointer(row.UnitID)
		tracking.UnitKey = row.UnitKey
		tracking.Amount = postgres.Float4Pointer(row.Amount)
		tracking.Modifier = postgres.Float4Pointer(row.Modifier)
	}

	return &Log{
		ID:              uuid.UUID(row.ID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		Description:     postgres.TextPointer(row.Description),
		LanguageCode:    row.LanguageCode,
		LanguageName:    row.LanguageName,
		Activity:        activities.Activity{ID: int32(row.ActivityID)},
		UnitID:          uuid.UUID(row.UnitID.Bytes),
		UnitKey:         row.UnitKey,
		UnitName:        row.UnitName,
		Tags:            legacyLogTags(row.Tags),
		Amount:          row.Amount.Float32,
		Modifier:        row.Modifier.Float32,
		Score:           row.Score.Float32,
		DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
		CreatedAt:       row.CreatedAt.Time,
		Deleted:         row.DeletedAt.Valid,
		UserDisplayName: &row.UserDisplayName,
		Tracking:        tracking,
	}, nil
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

func (r *LogsRepository) CanDelete(ctx context.Context, logID uuid.UUID, now time.Time) (bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}

	return queries.New(executor).CanDeleteLog(ctx, queries.CanDeleteLogParams{
		LogID: postgres.UUID(logID),
		Now:   pgtype.Date{Time: now, Valid: true},
	})
}

func (r *LogsRepository) AttachedContestIDs(ctx context.Context, logID uuid.UUID) ([]uuid.UUID, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListAttachedContestIDs(ctx, postgres.UUID(logID))
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		result[i] = row.Bytes
	}
	return result, nil
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

func (r *LogsRepository) ListUserLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLogsForUser(ctx, queries.ListLogsForUserParams{
		UserID:         postgres.NullableUUID(parameters.UserID),
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     postgres.TextPointer(row.Description),
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.ActivityID)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey,
			UnitName:        row.UnitName,
			Tags:            legacyLogTags(row.Tags),
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
		})
		result.TotalSize = int(row.TotalSize)
	}
	if parameters.Page*parameters.PageSize+parameters.PageSize < result.TotalSize {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

func (r *LogsRepository) ListContestLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLogsForContest(ctx, queries.ListLogsForContestParams{
		UserID:         postgres.NullableUUID(parameters.UserID),
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
		IncludeDeleted: parameters.IncludeDeleted,
		ContestID:      postgres.UUID(parameters.ContestID),
	})
	if err != nil {
		return nil, fmt.Errorf("list logs: %w", err)
	}

	result := &LogList{Logs: make([]Log, 0, len(rows))}
	for _, row := range rows {
		result.Logs = append(result.Logs, Log{
			ID:              uuid.UUID(row.ID.Bytes),
			UserID:          uuid.UUID(row.UserID.Bytes),
			Description:     postgres.TextPointer(row.Description),
			LanguageCode:    row.LanguageCode,
			LanguageName:    row.LanguageName,
			Activity:        activities.Activity{ID: int32(row.ActivityID)},
			UnitID:          uuid.UUID(row.UnitID.Bytes),
			UnitKey:         row.UnitKey,
			UnitName:        row.UnitName,
			Tags:            legacyLogTags(row.Tags),
			Amount:          row.Amount.Float32,
			Modifier:        row.Modifier.Float32,
			Score:           row.Score.Float32,
			DurationSeconds: postgres.Int4Pointer(row.DurationSeconds),
			CreatedAt:       row.CreatedAt.Time,
			Deleted:         row.DeletedAt.Valid,
			UserDisplayName: &row.UserDisplayName,
		})
		result.TotalSize = int(row.TotalSize)
	}
	if parameters.Page*parameters.PageSize+parameters.PageSize < result.TotalSize {
		result.NextPageToken = fmt.Sprint(parameters.Page + 1)
	}
	return result, nil
}

// Decode PostgreSQL array text at the storage boundary, preserving the legacy
// handling of escaped quotes and backslashes. Changing the decoded values
// requires a separate response-contract fix.
func legacyLogTags(encoded string) []string {
	result := []string{}
	if len(encoded) <= 2 {
		return result
	}
	var current []byte
	quoted := false
	for _, ch := range []byte(encoded[1 : len(encoded)-1]) {
		switch {
		case ch == '"':
			quoted = !quoted
		case ch == ',' && !quoted:
			result = append(result, string(current))
			current = nil
		default:
			current = append(current, ch)
		}
	}
	return append(result, string(current))
}
