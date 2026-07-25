package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestEvaluateScoringRuleSet(t *testing.T) {
	t.Run("uses the matching rule with the lowest priority number", func(t *testing.T) {
		amount := float32(100)
		laterRuleID := uuid.New()
		earlierRuleID := uuid.New()
		ruleSet := domain.ScoringRuleSet{
			ID:      uuid.New(),
			Version: 1,
			Rules: []domain.ScoringRule{
				{
					ID:          laterRuleID,
					Priority:    20,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1,
				},
				{
					ID:          earlierRuleID,
					Priority:    10,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        2,
				},
			},
		}

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID: 1,
			Amount:     &amount,
		}, ruleSet)

		require.NoError(t, err)
		assert.Equal(t, float32(200), result.Score)
		assert.Equal(t, earlierRuleID, *result.AppliedRuleID)
	})

	t.Run("requires every populated matcher to match", func(t *testing.T) {
		amount := float32(100)
		specificRuleID := uuid.New()
		fallbackRuleID := uuid.New()
		ruleSet := domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:           specificRuleID,
					Priority:     1,
					ActivityID:   1,
					UnitKey:      "reading_page",
					LanguageCode: "jpn",
					Tag:          "two_column",
					ScoreSource:  domain.ScoreSourceAmount,
					Rate:         1.6,
				},
				{
					ID:          fallbackRuleID,
					Priority:    2,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1,
				},
			},
		}

		matchingResult, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:   1,
			UnitKey:      "reading_page",
			LanguageCode: "jpn",
			Tags:         []string{"book", "two_column"},
			Amount:       &amount,
		}, ruleSet)
		require.NoError(t, err)
		assert.Equal(t, float32(160), matchingResult.Score)
		assert.Equal(t, specificRuleID, *matchingResult.AppliedRuleID)

		tests := []struct {
			name        string
			input       domain.ScoringInput
			wantMatched bool
		}{
			{
				name: "activity",
				input: domain.ScoringInput{
					ActivityID:   3,
					UnitKey:      "reading_page",
					LanguageCode: "jpn",
					Tags:         []string{"two_column"},
					Amount:       &amount,
				},
				wantMatched: false,
			},
			{
				name: "unit",
				input: domain.ScoringInput{
					ActivityID:   1,
					UnitKey:      "reading_sentence",
					LanguageCode: "jpn",
					Tags:         []string{"two_column"},
					Amount:       &amount,
				},
				wantMatched: true,
			},
			{
				name: "language",
				input: domain.ScoringInput{
					ActivityID:   1,
					UnitKey:      "reading_page",
					LanguageCode: "eng",
					Tags:         []string{"two_column"},
					Amount:       &amount,
				},
				wantMatched: true,
			},
			{
				name: "tag",
				input: domain.ScoringInput{
					ActivityID:   1,
					UnitKey:      "reading_page",
					LanguageCode: "jpn",
					Tags:         []string{"book"},
					Amount:       &amount,
				},
				wantMatched: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := domain.EvaluateScoringRuleSet(tt.input, ruleSet)

				require.NoError(t, err)
				if !tt.wantMatched {
					assert.False(t, result.Matched)
					return
				}
				assert.Equal(t, fallbackRuleID, *result.AppliedRuleID)
			})
		}
	})

	t.Run("calculates amount scores and returns provenance", func(t *testing.T) {
		amount := float32(200)
		ruleID := uuid.New()
		ruleSetID := uuid.New()

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID: 1,
			UnitKey:    "reading_page",
			Amount:     &amount,
		}, domain.ScoringRuleSet{
			ID: ruleSetID,
			Rules: []domain.ScoringRule{{
				ID:          ruleID,
				Priority:    1,
				ActivityID:  1,
				UnitKey:     "reading_page",
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        1,
			}},
		})

		require.NoError(t, err)
		assert.Equal(t, float32(200), result.Score)
		assert.Equal(t, domain.ScoreSourceAmount, result.ScoreSource)
		assert.True(t, result.Matched)
		assert.Equal(t, ruleSetID, *result.AppliedRuleSetID)
		assert.Equal(t, ruleID, *result.AppliedRuleID)
		assert.Equal(t, float32(1), *result.AppliedRate)
	})

	t.Run("converts duration from seconds to minutes", func(t *testing.T) {
		durationSeconds := int32(90)

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:      2,
			DurationSeconds: &durationSeconds,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        0.4,
			}},
		})

		require.NoError(t, err)
		assert.InDelta(t, 0.6, result.Score, 0.0001)
		assert.Equal(t, domain.ScoreSourceDurationMinutes, result.ScoreSource)
	})

	t.Run("amount takes precedence when amount and duration are both present", func(t *testing.T) {
		amount := float32(10)
		durationSeconds := int32(600)
		amountRuleID := uuid.New()

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:      1,
			Amount:          &amount,
			DurationSeconds: &durationSeconds,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:          uuid.New(),
					Priority:    1,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceDurationMinutes,
					Rate:        100,
				},
				{
					ID:          amountRuleID,
					Priority:    2,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        0.5,
				},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, float32(5), result.Score)
		assert.Equal(t, domain.ScoreSourceAmount, result.ScoreSource)
		assert.Equal(t, amountRuleID, *result.AppliedRuleID)
	})

	t.Run("returns zero and the selected source when no platform rule matches", func(t *testing.T) {
		durationSeconds := int32(60)

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:      5,
			DurationSeconds: &durationSeconds,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  1,
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        1,
			}},
		})

		require.NoError(t, err)
		assert.Zero(t, result.Score)
		assert.Equal(t, domain.ScoreSourceDurationMinutes, result.ScoreSource)
		assert.False(t, result.Matched)
		assert.Nil(t, result.AppliedRuleSetID)
		assert.Nil(t, result.AppliedRuleID)
		assert.Nil(t, result.AppliedRate)
	})
}

func TestEvaluateScoringRuleSetRejectsInvalidConfigurationAndInput(t *testing.T) {
	amount := float32(1)

	t.Run("duplicate priorities", func(t *testing.T) {
		_, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID: 1,
			Amount:     &amount,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:          uuid.New(),
					Priority:    1,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1,
				},
				{
					ID:          uuid.New(),
					Priority:    1,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        2,
				},
			},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidScoringRuleSet)
	})

	t.Run("missing tracking input", func(t *testing.T) {
		_, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID: 1,
		}, domain.ScoringRuleSet{ID: uuid.New()})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
	})
}

func TestEvaluateContestScore(t *testing.T) {
	amount := float32(100)
	input := domain.ScoringInput{
		ActivityID: 1,
		UnitKey:    "reading_page",
		Amount:     &amount,
	}
	platformRuleID := uuid.New()
	platformSet := domain.ScoringRuleSet{
		ID:      uuid.New(),
		Version: 3,
		Rules: []domain.ScoringRule{{
			ID:          platformRuleID,
			Priority:    1,
			ActivityID:  1,
			UnitKey:     "reading_page",
			ScoreSource: domain.ScoreSourceAmount,
			Rate:        1,
		}},
	}

	t.Run("override falls back to the pinned platform rule set", func(t *testing.T) {
		result, err := domain.EvaluateContestScore(input, domain.ScoringRuleSet{
			ID:      uuid.New(),
			Version: 1,
			Mode:    domain.ScoringRuleSetModeOverride,
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        2,
			}},
		}, &platformSet)

		require.NoError(t, err)
		assert.Equal(t, float32(100), result.Score)
		assert.Equal(t, platformRuleID, *result.AppliedRuleID)
		assert.Equal(t, platformSet.ID, *result.AppliedRuleSetID)
	})

	t.Run("replace returns zero for inputs its rules do not cover", func(t *testing.T) {
		result, err := domain.EvaluateContestScore(input, domain.ScoringRuleSet{
			ID:      uuid.New(),
			Version: 1,
			Mode:    domain.ScoringRuleSetModeReplace,
			Rules: []domain.ScoringRule{{
				ID:          uuid.New(),
				Priority:    1,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        2,
			}},
		}, &platformSet)

		require.NoError(t, err)
		assert.Zero(t, result.Score)
		assert.False(t, result.Matched)
	})

	t.Run("a pinned platform version remains stable after a newer version is published", func(t *testing.T) {
		newPlatformSet := platformSet
		newPlatformSet.ID = uuid.New()
		newPlatformSet.Version = 4
		newPlatformSet.Rules = append([]domain.ScoringRule(nil), platformSet.Rules...)
		newPlatformSet.Rules[0].ID = uuid.New()
		newPlatformSet.Rules[0].Rate = 2
		contestSet := domain.ScoringRuleSet{
			ID:      uuid.New(),
			Version: 1,
			Mode:    domain.ScoringRuleSetModeOverride,
			Rules:   []domain.ScoringRule{},
		}

		pinnedResult, err := domain.EvaluateContestScore(input, contestSet, &platformSet)
		require.NoError(t, err)
		newResult, err := domain.EvaluateContestScore(input, contestSet, &newPlatformSet)
		require.NoError(t, err)

		assert.Equal(t, float32(100), pinnedResult.Score)
		assert.Equal(t, platformSet.ID, *pinnedResult.AppliedRuleSetID)
		assert.Equal(t, float32(200), newResult.Score)
		assert.Equal(t, newPlatformSet.ID, *newResult.AppliedRuleSetID)
	})

	t.Run("override requires a pinned platform fallback", func(t *testing.T) {
		_, err := domain.EvaluateContestScore(input, domain.ScoringRuleSet{
			ID:      uuid.New(),
			Version: 1,
			Mode:    domain.ScoringRuleSetModeOverride,
		}, nil)

		assert.ErrorIs(t, err, domain.ErrInvalidScoringRuleSet)
	})
}
