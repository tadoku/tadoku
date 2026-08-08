package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

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
