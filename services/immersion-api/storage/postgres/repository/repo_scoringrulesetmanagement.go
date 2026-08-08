package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ListPlatformScoringRuleSets(ctx context.Context) ([]domain.ScoringRuleSet, error) {
	ruleSets, err := r.q.ListPlatformScoringRuleSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list platform scoring rule sets: %w", err)
	}
	result, err := r.loadScoringRuleSets(ctx, ruleSets)
	if err != nil {
		return nil, err
	}
	active, err := r.q.FindActivePlatformScoringRuleSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not resolve active platform scoring rule set: %w", err)
	}
	for i := range result {
		result[i].Active = result[i].ID == active.ID
	}
	return result, nil
}

func (r *Repository) ListContestScoringRuleSets(
	ctx context.Context,
	contestID uuid.UUID,
) ([]domain.ScoringRuleSet, error) {
	ruleSets, err := r.q.ListContestScoringRuleSets(ctx, uuid.NullUUID{UUID: contestID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("could not list contest scoring rule sets: %w", err)
	}
	result, err := r.loadScoringRuleSets(ctx, ruleSets)
	if err != nil {
		return nil, err
	}
	activeID, err := r.q.FindContestScoringRuleSetID(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve active contest scoring rule set: %w", err)
	}
	if activeID.Valid {
		for i := range result {
			result[i].Active = result[i].ID == activeID.UUID
		}
	}
	return result, nil
}

func (r *Repository) loadScoringRuleSets(
	ctx context.Context,
	rows []postgres.ScoringRuleSet,
) ([]domain.ScoringRuleSet, error) {
	result := make([]domain.ScoringRuleSet, len(rows))
	for i, row := range rows {
		ruleSet, err := loadScoringRuleSet(ctx, r.q, row)
		if err != nil {
			return nil, err
		}
		result[i] = *ruleSet
	}
	return result, nil
}

func (r *Repository) CreateScoringRuleSetDraft(
	ctx context.Context,
	req *domain.ScoringRuleSetDraftCreateRequest,
) (*domain.ScoringRuleSet, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not start scoring rule set transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	var version int32
	switch req.Scope() {
	case domain.ScoringRuleSetScopePlatform:
		version, err = qtx.NextPlatformScoringRuleSetVersion(ctx)
	case domain.ScoringRuleSetScopeContest:
		version, err = qtx.NextContestScoringRuleSetVersion(ctx, nullableUUID(req.ContestID))
	default:
		err = domain.ErrInvalidScoringRuleSet
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not allocate scoring rule set version: %w", err)
	}

	id := uuid.New()
	row, err := qtx.CreateScoringRuleSet(ctx, postgres.CreateScoringRuleSetParams{
		ID:                id,
		Scope:             string(req.Scope()),
		ContestID:         nullableUUID(req.ContestID),
		Version:           version,
		Mode:              nullableText(string(req.Mode)),
		FallbackRuleSetID: nullableUUID(req.FallbackRuleSetID),
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create scoring rule set: %w", err)
	}
	rules := make([]postgres.ScoringRule, len(req.Rules))
	for i, rule := range req.Rules {
		ruleID := uuid.New()
		if err := qtx.CreateScoringRule(ctx, postgres.CreateScoringRuleParams{
			ID:           ruleID,
			RuleSetID:    id,
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityID:   int16(rule.ActivityID),
			UnitKey:      nullableText(rule.UnitKey),
			LanguageCode: nullableText(rule.LanguageCode),
			Tag:          nullableText(rule.Tag),
			ScoreSource:  string(rule.ScoreSource),
			Rate:         rule.Rate,
		}); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("could not create scoring rule: %w", err)
		}
		rules[i] = postgres.ScoringRule{
			ID:           ruleID,
			RuleSetID:    id,
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityID:   int16(rule.ActivityID),
			UnitKey:      nullableText(rule.UnitKey),
			LanguageCode: nullableText(rule.LanguageCode),
			Tag:          nullableText(rule.Tag),
			ScoreSource:  string(rule.ScoreSource),
			Rate:         rule.Rate,
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not commit scoring rule set: %w", err)
	}
	return scoringRuleSetToDomain(row, rules), nil
}

func (r *Repository) PublishScoringRuleSet(
	ctx context.Context,
	id uuid.UUID,
	publishedAt time.Time,
) (*domain.ScoringRuleSet, error) {
	row, err := r.q.PublishScoringRuleSet(ctx, postgres.PublishScoringRuleSetParams{
		ID:          id,
		PublishedAt: postgres.NewNullTime(&publishedAt),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConflict
		}
		return nil, fmt.Errorf("could not publish scoring rule set: %w", err)
	}
	return loadScoringRuleSet(ctx, r.q, row)
}

func (r *Repository) ActivatePlatformScoringRuleSet(ctx context.Context, id uuid.UUID) error {
	if err := r.q.ActivatePlatformScoringRuleSet(ctx, id); err != nil {
		return fmt.Errorf("could not activate platform scoring rule set: %w", err)
	}
	return nil
}

func (r *Repository) ActivateContestScoringRuleSet(
	ctx context.Context,
	contestID uuid.UUID,
	ruleSetID uuid.UUID,
	updatedAt time.Time,
) error {
	if err := r.q.ActivateContestScoringRuleSet(ctx, postgres.ActivateContestScoringRuleSetParams{
		ContestID: contestID,
		RuleSetID: uuid.NullUUID{UUID: ruleSetID, Valid: true},
		UpdatedAt: updatedAt,
	}); err != nil {
		return fmt.Errorf("could not activate contest scoring rule set: %w", err)
	}
	return nil
}

func nullableUUID(value *uuid.UUID) uuid.NullUUID {
	if value == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *value, Valid: true}
}

func nullableText(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
