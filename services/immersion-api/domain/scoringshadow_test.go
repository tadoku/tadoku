package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockActivePlatformScoringRuleSetFinder struct {
	ruleSet *domain.ScoringRuleSet
	err     error
}

func (m *mockActivePlatformScoringRuleSetFinder) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	return m.ruleSet, m.err
}

func TestEvaluatePlatformScoringShadow(t *testing.T) {
	ruleSetID := uuid.New()
	ruleID := uuid.New()
	amount := float32(10)
	finder := &mockActivePlatformScoringRuleSetFinder{
		ruleSet: &domain.ScoringRuleSet{
			ID: ruleSetID,
			Rules: []domain.ScoringRule{{
				ID:          ruleID,
				Priority:    1,
				ActivityID:  1,
				UnitKey:     domain.UnitKeyReadingPage,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        1,
			}},
		},
	}

	t.Run("reports matching scores", func(t *testing.T) {
		comparison, err := domain.EvaluatePlatformScoringShadow(
			context.Background(),
			finder,
			domain.ScoringInput{
				ActivityID: 1,
				UnitKey:    domain.UnitKeyReadingPage,
				Amount:     &amount,
			},
			10,
		)

		require.NoError(t, err)
		assert.True(t, comparison.RuleResult.Matched)
		assert.False(t, comparison.Mismatch)
		assert.Equal(t, float32(10), comparison.RuleResult.Score)
	})

	t.Run("reports mismatched scores", func(t *testing.T) {
		comparison, err := domain.EvaluatePlatformScoringShadow(
			context.Background(),
			finder,
			domain.ScoringInput{
				ActivityID: 1,
				UnitKey:    domain.UnitKeyReadingPage,
				Amount:     &amount,
			},
			12,
		)

		require.NoError(t, err)
		assert.True(t, comparison.Mismatch)
	})

	t.Run("reports unmatched inputs without failing", func(t *testing.T) {
		comparison, err := domain.EvaluatePlatformScoringShadow(
			context.Background(),
			finder,
			domain.ScoringInput{
				ActivityID: 1,
				UnitKey:    domain.UnitKeyReadingSentence,
				Amount:     &amount,
			},
			0.5,
		)

		require.NoError(t, err)
		assert.False(t, comparison.RuleResult.Matched)
		assert.True(t, comparison.Mismatch)
		assert.Zero(t, comparison.RuleResult.Score)
	})

	t.Run("returns rule loading errors", func(t *testing.T) {
		loadErr := errors.New("database error")

		_, err := domain.EvaluatePlatformScoringShadow(
			context.Background(),
			&mockActivePlatformScoringRuleSetFinder{err: loadErr},
			domain.ScoringInput{ActivityID: 1, Amount: &amount},
			10,
		)

		assert.ErrorIs(t, err, loadErr)
	})
}

func TestEvaluatePlatformScoringShadowUsesDurationMinutes(t *testing.T) {
	duration := int32(90)
	finder := &mockActivePlatformScoringRuleSetFinder{
		ruleSet: &domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        0.4,
			}},
		},
	}

	comparison, err := domain.EvaluatePlatformScoringShadow(
		context.Background(),
		finder,
		domain.ScoringInput{
			ActivityID:      2,
			DurationSeconds: &duration,
		},
		0.6,
	)

	require.NoError(t, err)
	assert.False(t, comparison.Mismatch)
	assert.InDelta(t, float32(0.6), comparison.RuleResult.Score, 0.0001)
}
