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
	calls   int
}

func (m *mockActivePlatformScoringRuleSetFinder) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	m.calls++
	return m.ruleSet, m.err
}

type recordingScoringShadowObserver struct {
	observations []domain.ScoringShadowObservation
}

func (o *recordingScoringShadowObserver) ObserveScoringShadow(_ context.Context, observation domain.ScoringShadowObservation) {
	o.observations = append(o.observations, observation)
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

func TestEvaluateAndObservePlatformScoringRecordsOneMutuallyExclusiveOutcome(t *testing.T) {
	amount := float32(10)
	matchingRuleSet := func(rate float32) *domain.ScoringRuleSet {
		return &domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  1,
				UnitKey:     domain.UnitKeyReadingPage,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        rate,
			}},
		}
	}
	unmatchedRuleSet := &domain.ScoringRuleSet{
		ID: uuid.New(),
		Rules: []domain.ScoringRule{{
			ID:          uuid.New(),
			Priority:    1,
			ActivityID:  2,
			ScoreSource: domain.ScoreSourceAmount,
			Rate:        1,
		}},
	}

	testCases := []struct {
		name      string
		ruleSet   *domain.ScoringRuleSet
		err       error
		outcome   domain.ScoringShadowOutcome
		wantError bool
	}{
		{name: "match", ruleSet: matchingRuleSet(1), outcome: domain.ScoringShadowOutcomeMatch},
		{name: "mismatch", ruleSet: matchingRuleSet(2), outcome: domain.ScoringShadowOutcomeMismatch},
		{name: "unmatched", ruleSet: unmatchedRuleSet, outcome: domain.ScoringShadowOutcomeUnmatched},
		{name: "error", err: errors.New("database connection details must stay private"), outcome: domain.ScoringShadowOutcomeError, wantError: true},
	}
	modes := []domain.ScoringShadowMode{
		domain.ScoringShadowModeShadow,
		domain.ScoringShadowModeAuthoritative,
	}

	for _, mode := range modes {
		for _, testCase := range testCases {
			t.Run(string(mode)+"/"+testCase.name, func(t *testing.T) {
				finder := &mockActivePlatformScoringRuleSetFinder{ruleSet: testCase.ruleSet, err: testCase.err}
				observer := &recordingScoringShadowObserver{}

				_, err := domain.EvaluateAndObservePlatformScoring(
					context.Background(),
					finder,
					observer,
					domain.ScoringShadowOperationCreate,
					mode,
					domain.ScoringInput{
						ActivityID:   1,
						UnitKey:      domain.UnitKeyReadingPage,
						LanguageCode: "jpn",
						Amount:       &amount,
					},
					10,
				)

				if testCase.wantError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Equal(t, 1, finder.calls, "the engine must be evaluated exactly once")
				require.Len(t, observer.observations, 1, "one comparison must produce one outcome")
				observation := observer.observations[0]
				assert.Equal(t, testCase.outcome, observation.Outcome)
				assert.Equal(t, mode, observation.Mode)
				assert.Equal(t, domain.ScoringShadowOperationCreate, observation.Operation)
				assert.Equal(t, domain.ScoreSourceAmount, observation.ScoreSource)
				assert.Equal(t, int32(1), observation.ActivityID)
				if testCase.wantError {
					assert.Equal(t, "evaluation_failed", observation.ErrorType)
					assert.NotContains(t, observation.ErrorType, "database")
				} else {
					require.NotNil(t, observation.EngineScore)
					require.NotNil(t, observation.AbsoluteDelta)
					require.NotNil(t, observation.RelativeDelta)
				}
			})
		}
	}
}
