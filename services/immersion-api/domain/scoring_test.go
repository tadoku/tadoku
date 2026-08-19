package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestEvaluateScoringRuleSet(t *testing.T) {
	t.Run("uses the most specific matching base rule", func(t *testing.T) {
		amount := float32(100)
		laterRuleID := uuid.New()
		earlierRuleID := uuid.New()
		ruleSet := domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:          laterRuleID,
					Priority:    20,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        2,
				},
				{
					ID:           earlierRuleID,
					Priority:     10,
					ActivityID:   1,
					LanguageCode: "jpn",
					ScoreSource:  domain.ScoreSourceAmount,
					Rate:         1.5,
				},
			},
		}

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount,
		}, ruleSet)

		require.NoError(t, err)
		assert.Equal(t, float32(150), result.Score)
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: earlierRuleID, Rate: 1.5},
		}, result.AppliedRules)
	})

	t.Run("stacks every matching modifier after the selected base rule", func(t *testing.T) {
		amount := float32(100)
		baseRuleID := uuid.New()
		earlierModifierID := uuid.New()
		laterModifierID := uuid.New()

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID:   1,
			LanguageCode: "jpn",
			Tags:         []string{"dense", "featured"},
			Amount:       &amount,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:           laterModifierID,
					Priority:     30,
					Stackable:    true,
					ActivityID:   1,
					LanguageCode: "jpn",
					Tag:          "featured",
					ScoreSource:  domain.ScoreSourceAmount,
					Rate:         3,
				},
				{
					ID:          baseRuleID,
					Priority:    10,
					ActivityID:  1,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1.5,
				},
				{
					ID:          earlierModifierID,
					Priority:    5,
					Stackable:   true,
					ActivityID:  1,
					Tag:         "dense",
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        2,
				},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, float32(900), result.Score)
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: baseRuleID, Rate: 1.5},
			{RuleID: earlierModifierID, Rate: 2},
			{RuleID: laterModifierID, Rate: 3},
		}, result.AppliedRules)
	})

	t.Run("does not match modifiers without a base rule", func(t *testing.T) {
		amount := float32(100)

		result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
			ActivityID: 1,
			Tags:       []string{"dense"},
			Amount:     &amount,
		}, domain.ScoringRuleSet{
			ID: uuid.New(),
			Rules: []domain.ScoringRule{
				{
					ID:          uuid.New(),
					Priority:    1,
					Stackable:   true,
					ActivityID:  1,
					Tag:         "dense",
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1.5,
				},
				{
					ID:          uuid.New(),
					Priority:    2,
					ActivityID:  2,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        0.5,
				},
			},
		})

		require.NoError(t, err)
		assert.Zero(t, result.Score)
		assert.False(t, result.Matched)
		assert.Nil(t, result.AppliedRuleSetID)
		assert.Empty(t, result.AppliedRules)
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
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: specificRuleID, Rate: 1.6},
		}, matchingResult.AppliedRules)

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
				assert.Equal(t, []domain.AppliedScoringRule{
					{RuleID: fallbackRuleID, Rate: 1},
				}, result.AppliedRules)
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
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: ruleID, Rate: 1},
		}, result.AppliedRules)
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
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: amountRuleID, Rate: 0.5},
		}, result.AppliedRules)
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
		assert.Empty(t, result.AppliedRules)
	})
}

func TestDenseListeningTagAppliesOnlyToDurationMinutes(t *testing.T) {
	amountBaseRuleID := uuid.New()
	legacyDenseRuleID := uuid.New()
	durationBaseRuleID := uuid.New()
	denseTagRuleID := uuid.New()
	ruleSet := domain.ScoringRuleSet{
		ID: uuid.New(),
		Rules: []domain.ScoringRule{
			{
				ID:          amountBaseRuleID,
				Priority:    200,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        0.4,
			},
			{
				ID:          legacyDenseRuleID,
				Priority:    210,
				Stackable:   true,
				ActivityID:  2,
				UnitKey:     domain.UnitKeyListeningDenseMinutes,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        1.5,
			},
			{
				ID:          denseTagRuleID,
				Priority:    220,
				Stackable:   true,
				ActivityID:  2,
				Tag:         "dense",
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        1.5,
			},
			{
				ID:          durationBaseRuleID,
				Priority:    610,
				ActivityID:  2,
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        0.4,
			},
		},
	}
	amount := float32(60)
	durationSeconds := int32(3600)

	tests := []struct {
		name             string
		input            domain.ScoringInput
		wantScore        float32
		wantSource       domain.ScoreSource
		wantAppliedRules []domain.AppliedScoringRule
	}{
		{
			name: "plain duration minutes keep the base rate",
			input: domain.ScoringInput{
				ActivityID:      2,
				DurationSeconds: &durationSeconds,
			},
			wantScore:  24,
			wantSource: domain.ScoreSourceDurationMinutes,
			wantAppliedRules: []domain.AppliedScoringRule{
				{RuleID: durationBaseRuleID, Rate: 0.4},
			},
		},
		{
			name: "dense tag boosts duration minutes",
			input: domain.ScoringInput{
				ActivityID:      2,
				Tags:            []string{"dense"},
				DurationSeconds: &durationSeconds,
			},
			wantScore:  36,
			wantSource: domain.ScoreSourceDurationMinutes,
			wantAppliedRules: []domain.AppliedScoringRule{
				{RuleID: durationBaseRuleID, Rate: 0.4},
				{RuleID: denseTagRuleID, Rate: 1.5},
			},
		},
		{
			name: "legacy dense-minute amount keeps its existing rate",
			input: domain.ScoringInput{
				ActivityID: 2,
				UnitKey:    domain.UnitKeyListeningDenseMinutes,
				Tags:       []string{"dense"},
				Amount:     &amount,
			},
			wantScore:  36,
			wantSource: domain.ScoreSourceAmount,
			wantAppliedRules: []domain.AppliedScoringRule{
				{RuleID: amountBaseRuleID, Rate: 0.4},
				{RuleID: legacyDenseRuleID, Rate: 1.5},
			},
		},
		{
			name: "dense tag does not boost legacy normal-minute amounts",
			input: domain.ScoringInput{
				ActivityID: 2,
				UnitKey:    domain.UnitKeyListeningMinute,
				Tags:       []string{"dense"},
				Amount:     &amount,
			},
			wantScore:  24,
			wantSource: domain.ScoreSourceAmount,
			wantAppliedRules: []domain.AppliedScoringRule{
				{RuleID: amountBaseRuleID, Rate: 0.4},
			},
		},
		{
			name: "amount remains authoritative when both inputs are present",
			input: domain.ScoringInput{
				ActivityID:      2,
				UnitKey:         domain.UnitKeyListeningDenseMinutes,
				Tags:            []string{"dense"},
				Amount:          &amount,
				DurationSeconds: &durationSeconds,
			},
			wantScore:  36,
			wantSource: domain.ScoreSourceAmount,
			wantAppliedRules: []domain.AppliedScoringRule{
				{RuleID: amountBaseRuleID, Rate: 0.4},
				{RuleID: legacyDenseRuleID, Rate: 1.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := domain.EvaluateScoringRuleSet(tt.input, ruleSet)

			require.NoError(t, err)
			assert.InDelta(t, tt.wantScore, result.Score, 0.0001)
			assert.Equal(t, tt.wantSource, result.ScoreSource)
			assert.Equal(t, tt.wantAppliedRules, result.AppliedRules)
		})
	}
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

func TestPlatformRuleSetV2MatchesProductionLegacyUnitMatrix(t *testing.T) {
	type legacyUnit struct {
		name         string
		activityID   int32
		unitKey      string
		languageCode string
		modifier     float32
	}

	legacyUnits := []legacyUnit{
		{name: "reading/page", activityID: 1, unitKey: domain.UnitKeyReadingPage, modifier: 1},
		{name: "reading/two-column-page/jpn", activityID: 1, unitKey: domain.UnitKeyReadingTwoColumnPage, languageCode: "jpn", modifier: 1.6},
		{name: "reading/comic-page", activityID: 1, unitKey: domain.UnitKeyReadingComicPage, modifier: 0.2},
		{name: "reading/sentence", activityID: 1, unitKey: domain.UnitKeyReadingSentence, modifier: 0.05},
		{name: "reading/character/default", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, modifier: 0.000833333},
		{name: "reading/character/jpn", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "jpn", modifier: 0.0025},
		{name: "reading/character/kor", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "kor", modifier: 0.0025},
		{name: "reading/character/zho", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "zho", modifier: 0.0025},
		{name: "reading/character/cmn", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "cmn", modifier: 0.0025},
		{name: "reading/character/yue", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "yue", modifier: 0.0025},
		{name: "reading/character/wuu", activityID: 1, unitKey: domain.UnitKeyReadingCharacter, languageCode: "wuu", modifier: 0.0025},
		{name: "listening/minute", activityID: 2, unitKey: domain.UnitKeyListeningMinute, modifier: 0.4},
		{name: "listening/dense-minute", activityID: 2, unitKey: domain.UnitKeyListeningDenseMinutes, modifier: 0.6},
		{name: "writing/page", activityID: 3, unitKey: domain.UnitKeyWritingPage, modifier: 10},
		{name: "writing/sentence", activityID: 3, unitKey: domain.UnitKeyWritingSentence, modifier: 0.5},
		{name: "writing/character/default", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, modifier: 0.00833333},
		{name: "writing/character/jpn", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "jpn", modifier: 0.025},
		{name: "writing/character/kor", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "kor", modifier: 0.025},
		{name: "writing/character/zho", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "zho", modifier: 0.025},
		{name: "writing/character/cmn", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "cmn", modifier: 0.025},
		{name: "writing/character/yue", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "yue", modifier: 0.025},
		{name: "writing/character/wuu", activityID: 3, unitKey: domain.UnitKeyWritingCharacter, languageCode: "wuu", modifier: 0.025},
		{name: "speaking/minute", activityID: 4, unitKey: domain.UnitKeySpeakingMinute, modifier: 0.5},
		{name: "speaking/dense-minute", activityID: 4, unitKey: domain.UnitKeySpeakingDenseMinutes, modifier: 0.7},
		{name: "study/minute", activityID: 5, unitKey: domain.UnitKeyStudyMinute, modifier: 0.5},
	}

	newRule := func(priority, activityID int32, unitKey, languageCode string, rate float32, stackable bool) domain.ScoringRule {
		return domain.ScoringRule{
			ID:           uuid.New(),
			Priority:     priority,
			Stackable:    stackable,
			ActivityID:   activityID,
			UnitKey:      unitKey,
			LanguageCode: languageCode,
			ScoreSource:  domain.ScoreSourceAmount,
			Rate:         rate,
		}
	}
	highRateCharacterLanguages := []string{"jpn", "kor", "zho", "cmn", "yue", "wuu"}
	rules := []domain.ScoringRule{
		newRule(100, 1, domain.UnitKeyReadingTwoColumnPage, "jpn", 1.6, false),
		newRule(110, 1, domain.UnitKeyReadingPage, "", 1, false),
		newRule(120, 1, domain.UnitKeyReadingComicPage, "", 0.2, false),
		newRule(130, 1, domain.UnitKeyReadingSentence, "", 0.05, false),
	}
	for i, languageCode := range highRateCharacterLanguages {
		rules = append(rules, newRule(int32(140+i), 1, domain.UnitKeyReadingCharacter, languageCode, 0.0025, false))
	}
	rules = append(rules,
		newRule(150, 1, domain.UnitKeyReadingCharacter, "", 0.000833333, false),
		newRule(200, 2, "", "", 0.4, false),
		newRule(210, 2, domain.UnitKeyListeningDenseMinutes, "", 1.5, true),
		newRule(300, 3, domain.UnitKeyWritingPage, "", 10, false),
		newRule(310, 3, domain.UnitKeyWritingSentence, "", 0.5, false),
	)
	for i, languageCode := range highRateCharacterLanguages {
		rules = append(rules, newRule(int32(320+i), 3, domain.UnitKeyWritingCharacter, languageCode, 0.025, false))
	}
	rules = append(rules,
		newRule(330, 3, domain.UnitKeyWritingCharacter, "", 0.00833333, false),
		newRule(400, 4, "", "", 0.5, false),
		newRule(410, 4, domain.UnitKeySpeakingDenseMinutes, "", 1.4, true),
		newRule(500, 5, domain.UnitKeyStudyMinute, "", 0.5, false),
	)
	ruleSet := domain.ScoringRuleSet{ID: uuid.New(), Version: 2, Rules: rules}
	amounts := []float32{1, 80, 80.25}

	for _, unit := range legacyUnits {
		for _, amount := range amounts {
			t.Run(unit.name, func(t *testing.T) {
				result, err := domain.EvaluateScoringRuleSet(domain.ScoringInput{
					ActivityID:   unit.activityID,
					UnitKey:      unit.unitKey,
					LanguageCode: unit.languageCode,
					Amount:       &amount,
				}, ruleSet)

				require.NoError(t, err)
				require.True(t, result.Matched)
				assert.InDelta(t, amount*unit.modifier, result.Score, 0.000001)
			})
		}
	}
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
		assert.Equal(t, []domain.AppliedScoringRule{
			{RuleID: platformRuleID, Rate: 1},
		}, result.AppliedRules)
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
