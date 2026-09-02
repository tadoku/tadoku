package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres/postgrestest"
)

func TestActivePlatformScoringRulesSupportReadingPageTags(t *testing.T) {
	testDB := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(testDB)
	ruleSet, err := repo.FindActivePlatformScoringRuleSet(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ruleSet)
	assert.Equal(t, int32(4), ruleSet.Version)

	amount := float32(10)
	tests := []struct {
		name         string
		unitKey      string
		languageCode string
		tags         []string
		wantScore    float32
	}{
		{name: "ordinary pages", unitKey: domain.UnitKeyReadingPage, languageCode: "jpn", wantScore: 10},
		{name: "Japanese two-column page tag", unitKey: domain.UnitKeyReadingPage, languageCode: "jpn", tags: []string{"two_column"}, wantScore: 16},
		{name: "English two-column tag remains informational", unitKey: domain.UnitKeyReadingPage, languageCode: "eng", tags: []string{"two_column"}, wantScore: 10},
		{name: "comic page tag", unitKey: domain.UnitKeyReadingPage, languageCode: "eng", tags: []string{"comic"}, wantScore: 2},
		{name: "manga page tag", unitKey: domain.UnitKeyReadingPage, languageCode: "jpn", tags: []string{"manga"}, wantScore: 2},
		{name: "legacy two-column page unit", unitKey: domain.UnitKeyReadingTwoColumnPage, languageCode: "jpn", wantScore: 16},
		{name: "legacy comic page unit", unitKey: domain.UnitKeyReadingComicPage, languageCode: "eng", wantScore: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, evaluateErr := domain.EvaluateScoringRuleSet(domain.ScoringInput{
				ActivityID:   1,
				UnitKey:      tt.unitKey,
				LanguageCode: tt.languageCode,
				Tags:         tt.tags,
				Amount:       &amount,
			}, *ruleSet)

			require.NoError(t, evaluateErr)
			assert.Equal(t, tt.wantScore, result.Score)
		})
	}
}

func TestActivePlatformScoringRulesSupportDenseSpeakingDuration(t *testing.T) {
	testDB := postgrestest.OpenMigratedDatabase(t)
	repo := NewRepository(testDB)
	ruleSet, err := repo.FindActivePlatformScoringRuleSet(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ruleSet)
	assert.Equal(t, int32(4), ruleSet.Version)

	durationSeconds := int32(600)
	tests := []struct {
		name      string
		tags      []string
		wantScore float32
	}{
		{name: "plain speaking", wantScore: 5},
		{name: "dense speaking", tags: []string{"dense"}, wantScore: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, evaluateErr := domain.EvaluateScoringRuleSet(domain.ScoringInput{
				ActivityID:      4,
				Tags:            tt.tags,
				DurationSeconds: &durationSeconds,
			}, *ruleSet)

			require.NoError(t, evaluateErr)
			assert.InDelta(t, tt.wantScore, result.Score, 0.0001)
		})
	}
}

func TestScoringRuleSetToDomain(t *testing.T) {
	ruleSetID := uuid.New()
	contestID := uuid.New()
	fallbackID := uuid.New()
	baseRuleID := uuid.New()
	modifierRuleID := uuid.New()
	createdAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	publishedAt := createdAt.Add(time.Hour)

	result := scoringRuleSetToDomain(
		postgres.ScoringRuleSet{
			ID:                ruleSetID,
			Scope:             "contest",
			ContestID:         uuid.NullUUID{UUID: contestID, Valid: true},
			Version:           2,
			Status:            "published",
			Mode:              sql.NullString{String: "override", Valid: true},
			FallbackRuleSetID: uuid.NullUUID{UUID: fallbackID, Valid: true},
			CreatedAt:         createdAt,
			PublishedAt:       sql.NullTime{Time: publishedAt, Valid: true},
		},
		[]postgres.ScoringRule{
			{
				ID:           baseRuleID,
				Priority:     10,
				ActivityID:   1,
				UnitKey:      sql.NullString{String: domain.UnitKeyReadingPage, Valid: true},
				LanguageCode: sql.NullString{String: "jpn", Valid: true},
				ScoreSource:  "amount",
				Rate:         1.6,
			},
			{
				ID:          modifierRuleID,
				Priority:    20,
				Stackable:   true,
				ActivityID:  2,
				Tag:         sql.NullString{String: "dense", Valid: true},
				ScoreSource: "duration_minutes",
				Rate:        1.5,
			},
		},
	)

	require.NotNil(t, result)
	assert.Equal(t, ruleSetID, result.ID)
	assert.Equal(t, domain.ScoringRuleSetScopeContest, result.Scope)
	assert.Equal(t, &contestID, result.ContestID)
	assert.Equal(t, int32(2), result.Version)
	assert.Equal(t, domain.ScoringRuleSetStatusPublished, result.Status)
	assert.Equal(t, domain.ScoringRuleSetModeOverride, result.Mode)
	require.NotNil(t, result.FallbackRuleSetID)
	assert.Equal(t, fallbackID, *result.FallbackRuleSetID)
	assert.Equal(t, createdAt, result.CreatedAt)
	assert.Equal(t, &publishedAt, result.PublishedAt)
	require.Len(t, result.Rules, 2)
	assert.Equal(t, domain.UnitKeyReadingPage, result.Rules[0].UnitKey)
	assert.Equal(t, "jpn", result.Rules[0].LanguageCode)
	assert.Equal(t, domain.ScoreSourceAmount, result.Rules[0].ScoreSource)
	assert.True(t, result.Rules[1].Stackable)
	assert.Equal(t, "dense", result.Rules[1].Tag)
	assert.Equal(t, domain.ScoreSourceDurationMinutes, result.Rules[1].ScoreSource)
}

func TestScoringRuleSetToDomainLeavesOptionalFieldsEmpty(t *testing.T) {
	result := scoringRuleSetToDomain(
		postgres.ScoringRuleSet{ID: uuid.New(), Version: 1},
		[]postgres.ScoringRule{{
			ID:          uuid.New(),
			Priority:    1,
			ActivityID:  1,
			ScoreSource: "amount",
			Rate:        1,
		}},
	)

	assert.Empty(t, result.Mode)
	assert.Nil(t, result.FallbackRuleSetID)
	require.Len(t, result.Rules, 1)
	assert.Empty(t, result.Rules[0].UnitKey)
	assert.Empty(t, result.Rules[0].LanguageCode)
	assert.Empty(t, result.Rules[0].Tag)
}
