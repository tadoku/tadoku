package scoring

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type ScoringRepository struct{ db *pgxpool.Pool }

func NewScoringRepository(db *pgxpool.Pool) *ScoringRepository { return &ScoringRepository{db: db} }

func (r *ScoringRepository) ListPlatformRuleSets(ctx context.Context) ([]RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.ListPlatformScoringRuleSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list platform scoring rule sets: %w", err)
	}
	return ruleSets(rows), nil
}

func (r *ScoringRepository) ListContestRuleSets(ctx context.Context, contestID uuid.UUID) ([]RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.ListContestScoringRuleSets(ctx, postgres.UUID(contestID))
	if err != nil {
		return nil, fmt.Errorf("list contest scoring rule sets: %w", err)
	}
	return ruleSets(rows), nil
}

func (r *ScoringRepository) FindActivePlatformRuleSet(ctx context.Context) (*RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	row, err := q.FindActivePlatformScoringRuleSet(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleSetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find active platform scoring rule set: %w", err)
	}
	return ruleSet(row), nil
}

func (r *ScoringRepository) FindContestActiveRuleSetID(ctx context.Context, contestID uuid.UUID) (*uuid.UUID, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	id, err := q.FindContestScoringRuleSetID(ctx, postgres.UUID(contestID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find contest scoring rule set: %w", err)
	}
	if !id.Valid {
		return nil, nil
	}
	value := uuid.UUID(id.Bytes)
	return &value, nil
}

func (r *ScoringRepository) FindRuleSetByID(ctx context.Context, id uuid.UUID) (*RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	row, err := q.FindScoringRuleSetByID(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleSetNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find scoring rule set: %w", err)
	}
	return ruleSet(row), nil
}

func (r *ScoringRepository) ListRules(ctx context.Context, ruleSetID uuid.UUID) ([]Rule, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.ListScoringRulesForRuleSet(ctx, postgres.UUID(ruleSetID))
	if err != nil {
		return nil, fmt.Errorf("list scoring rules: %w", err)
	}
	rules := make([]Rule, len(rows))
	for i, row := range rows {
		rules[i] = Rule{
			ID:           row.ID.Bytes,
			Priority:     row.Priority,
			Stackable:    row.Stackable,
			ActivityID:   int32(row.ActivityID),
			UnitKey:      row.UnitKey.String,
			LanguageCode: row.LanguageCode.String,
			Tag:          row.Tag.String,
			Source:       Source(row.ScoreSource),
			Rate:         row.Rate,
		}
	}
	return rules, nil
}

func (r *ScoringRepository) FindUnitKeyByID(ctx context.Context, id uuid.UUID, activityID int32, languageCode string) (string, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return "", err
	}

	row, err := q.FindUnitForScoringByID(ctx, queries.FindUnitForScoringByIDParams{
		ID:           postgres.UUID(id),
		ActivityID:   int16(activityID),
		LanguageCode: pgtype.Text{String: languageCode, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errx.NewInvalidInputError("unit_id is not valid for activity_id and language_code")
	}
	if err != nil {
		return "", fmt.Errorf("find unit for scoring: %w", err)
	}
	return row.UnitKey, nil
}

func (r *ScoringRepository) FindUnitKeyByKey(ctx context.Context, key string, activityID int32, languageCode string) (string, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return "", err
	}

	row, err := q.FindUnitForScoringByKey(ctx, queries.FindUnitForScoringByKeyParams{
		UnitKey:      key,
		ActivityID:   int16(activityID),
		LanguageCode: pgtype.Text{String: languageCode, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errx.NewInvalidInputError("unit_key is not valid for activity_id and language_code")
	}
	if err != nil {
		return "", fmt.Errorf("find unit for scoring: %w", err)
	}
	return row.UnitKey, nil
}

type logUnit struct {
	ID       uuid.UUID
	Key      string
	Modifier float32
}

func (r *ScoringRepository) FindLogUnit(ctx context.Context, id *uuid.UUID, key *string, activityID int32, languageCode string) (*logUnit, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	if id != nil {
		row, err := q.FindUnitForScoringByID(ctx, queries.FindUnitForScoringByIDParams{
			ID:           postgres.UUID(*id),
			ActivityID:   int16(activityID),
			LanguageCode: pgtype.Text{String: languageCode, Valid: true},
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errx.NewInvalidInputError("unit_id is not valid")
		}
		if err != nil {
			return nil, fmt.Errorf("find log unit: %w", err)
		}
		return &logUnit{ID: row.ID.Bytes, Key: row.UnitKey, Modifier: row.Modifier}, nil
	}
	if key != nil {
		row, err := q.FindUnitForScoringByKey(ctx, queries.FindUnitForScoringByKeyParams{
			UnitKey:      *key,
			ActivityID:   int16(activityID),
			LanguageCode: pgtype.Text{String: languageCode, Valid: true},
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errx.NewInvalidInputError("unit_key is not valid")
		}
		if err != nil {
			return nil, fmt.Errorf("find log unit: %w", err)
		}
		return &logUnit{ID: row.ID.Bytes, Key: row.UnitKey, Modifier: row.Modifier}, nil
	}
	return nil, nil
}

func (r *ScoringRepository) NextDraftVersion(ctx context.Context, contestID *uuid.UUID) (int32, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return 0, err
	}
	if contestID == nil {
		return q.NextPlatformScoringRuleSetVersion(ctx)
	}
	return q.NextContestScoringRuleSetVersion(ctx, postgres.UUID(*contestID))
}

func (r *ScoringRepository) CreateDraft(ctx context.Context, draft RuleSet) (*RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	row, err := q.CreateScoringRuleSet(ctx, queries.CreateScoringRuleSetParams{
		ID:                postgres.UUID(draft.ID),
		Scope:             draft.Scope,
		ContestID:         postgres.NullableUUID(draft.ContestID),
		Version:           draft.Version,
		Mode:              postgres.NullableNonEmptyText(&draft.Mode),
		FallbackRuleSetID: postgres.NullableUUID(draft.FallbackRuleSetID),
		CreatedAt:         pgtype.Timestamp{Time: draft.CreatedAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create scoring rule set: %w", err)
	}
	return ruleSet(row), nil
}

func (r *ScoringRepository) CreateRule(ctx context.Context, ruleSetID uuid.UUID, rule Rule) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	if err := q.CreateScoringRule(ctx, queries.CreateScoringRuleParams{
		ID:           postgres.UUID(rule.ID),
		RuleSetID:    postgres.UUID(ruleSetID),
		Priority:     rule.Priority,
		Stackable:    rule.Stackable,
		ActivityID:   int16(rule.ActivityID),
		UnitKey:      postgres.NullableNonEmptyText(&rule.UnitKey),
		LanguageCode: postgres.NullableNonEmptyText(&rule.LanguageCode),
		Tag:          postgres.NullableNonEmptyText(&rule.Tag),
		ScoreSource:  string(rule.Source),
		Rate:         rule.Rate,
	}); err != nil {
		return fmt.Errorf("create scoring rule: %w", err)
	}
	return nil
}

func (r *ScoringRepository) PublishRuleSet(ctx context.Context, id uuid.UUID, publishedAt time.Time) (*RuleSet, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	row, err := q.PublishScoringRuleSet(ctx, queries.PublishScoringRuleSetParams{
		ID:          postgres.UUID(id),
		PublishedAt: pgtype.Timestamp{Time: publishedAt, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errx.NewConflictError("only draft scoring rule sets can be published")
	}
	if err != nil {
		return nil, fmt.Errorf("publish scoring rule set: %w", err)
	}
	return ruleSet(row), nil
}

func (r *ScoringRepository) ActivatePlatformRuleSet(ctx context.Context, id uuid.UUID) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	if err := q.ActivatePlatformScoringRuleSet(ctx, postgres.UUID(id)); err != nil {
		return fmt.Errorf("activate platform scoring rule set: %w", err)
	}
	return nil
}

func (r *ScoringRepository) ActivateContestRuleSet(ctx context.Context, contestID, id uuid.UUID, updatedAt time.Time) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	if err := q.ActivateContestScoringRuleSet(ctx, queries.ActivateContestScoringRuleSetParams{
		RuleSetID: postgres.UUID(id),
		UpdatedAt: pgtype.Timestamp{Time: updatedAt, Valid: true},
		ContestID: postgres.UUID(contestID),
	}); err != nil {
		return fmt.Errorf("activate contest scoring rule set: %w", err)
	}
	return nil
}

func (r *ScoringRepository) queries(ctx context.Context) (*queries.Queries, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return queries.New(executor), nil
}

func ruleSets(rows []queries.ScoringRuleSet) []RuleSet {
	result := make([]RuleSet, len(rows))
	for i, row := range rows {
		result[i] = *ruleSet(row)
	}
	return result
}

func ruleSet(row queries.ScoringRuleSet) *RuleSet {
	result := &RuleSet{
		ID:        row.ID.Bytes,
		Scope:     row.Scope,
		Version:   row.Version,
		Status:    row.Status,
		Mode:      row.Mode.String,
		Rules:     []Rule{},
		CreatedAt: row.CreatedAt.Time,
	}
	if row.ContestID.Valid {
		id := uuid.UUID(row.ContestID.Bytes)
		result.ContestID = &id
	}
	if row.FallbackRuleSetID.Valid {
		id := uuid.UUID(row.FallbackRuleSetID.Bytes)
		result.FallbackRuleSetID = &id
	}
	if row.PublishedAt.Valid {
		publishedAt := row.PublishedAt.Time
		result.PublishedAt = &publishedAt
	}
	return result
}
