package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FindActivePlatformScoringRuleSet(ctx context.Context) (*domain.ScoringRuleSet, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch active platform scoring rule set: %w", err)
	}
	qtx := r.q.WithTx(tx)

	ruleSet, err := qtx.FindActivePlatformScoringRuleSet(ctx)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrScoringRuleSetNotFound
		}

		return nil, fmt.Errorf("could not fetch active platform scoring rule set: %w", err)
	}

	result, err := loadScoringRuleSet(ctx, qtx, ruleSet)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not fetch active platform scoring rule set: %w", err)
	}

	return result, nil
}

func (r *Repository) FindScoringRuleSetByID(ctx context.Context, id uuid.UUID) (*domain.ScoringRuleSet, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scoring rule set: %w", err)
	}
	qtx := r.q.WithTx(tx)

	ruleSet, err := qtx.FindScoringRuleSetByID(ctx, id)
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrScoringRuleSetNotFound
		}

		return nil, fmt.Errorf("could not fetch scoring rule set: %w", err)
	}

	result, err := loadScoringRuleSet(ctx, qtx, ruleSet)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not fetch scoring rule set: %w", err)
	}

	return result, nil
}

func (r *Repository) FindContestScoringRuleSets(
	ctx context.Context,
	contestID uuid.UUID,
) (*domain.ScoringRuleSet, *domain.ScoringRuleSet, error) {
	ruleSetID, err := r.q.FindContestScoringRuleSetID(ctx, contestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, domain.ErrNotFound
		}

		return nil, nil, fmt.Errorf("could not fetch contest scoring rule set: %w", err)
	}
	if !ruleSetID.Valid {
		return nil, nil, nil
	}

	ruleSet, err := r.FindScoringRuleSetByID(ctx, ruleSetID.UUID)
	if err != nil {
		return nil, nil, err
	}
	if ruleSet.FallbackRuleSetID == nil {
		return ruleSet, nil, nil
	}

	fallback, err := r.FindScoringRuleSetByID(ctx, *ruleSet.FallbackRuleSetID)
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch contest fallback scoring rule set: %w", err)
	}

	return ruleSet, fallback, nil
}

func loadScoringRuleSet(
	ctx context.Context,
	queries *postgres.Queries,
	ruleSet postgres.ScoringRuleSet,
) (*domain.ScoringRuleSet, error) {
	rules, err := queries.ListScoringRulesForRuleSet(ctx, ruleSet.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scoring rules: %w", err)
	}

	return scoringRuleSetToDomain(ruleSet, rules), nil
}

func scoringRuleSetToDomain(
	ruleSet postgres.ScoringRuleSet,
	rules []postgres.ScoringRule,
) *domain.ScoringRuleSet {
	result := &domain.ScoringRuleSet{
		ID:        ruleSet.ID,
		Scope:     domain.ScoringRuleSetScope(ruleSet.Scope),
		Version:   ruleSet.Version,
		Status:    domain.ScoringRuleSetStatus(ruleSet.Status),
		Mode:      domain.ScoringRuleSetMode(ruleSet.Mode.String),
		Rules:     make([]domain.ScoringRule, len(rules)),
		CreatedAt: ruleSet.CreatedAt,
	}
	if ruleSet.ContestID.Valid {
		contestID := ruleSet.ContestID.UUID
		result.ContestID = &contestID
	}
	if ruleSet.FallbackRuleSetID.Valid {
		fallbackID := ruleSet.FallbackRuleSetID.UUID
		result.FallbackRuleSetID = &fallbackID
	}
	if ruleSet.PublishedAt.Valid {
		publishedAt := ruleSet.PublishedAt.Time
		result.PublishedAt = &publishedAt
	}
	for i, rule := range rules {
		result.Rules[i] = domain.ScoringRule{
			ID:           rule.ID,
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityID:   int32(rule.ActivityID),
			UnitKey:      rule.UnitKey.String,
			LanguageCode: rule.LanguageCode.String,
			Tag:          rule.Tag.String,
			ScoreSource:  domain.ScoreSource(rule.ScoreSource),
			Rate:         rule.Rate,
		}
	}
	return result
}
