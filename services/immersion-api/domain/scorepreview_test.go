package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockScorePreviewRepository struct {
	registrations   *domain.ContestRegistrations
	platformRuleSet *domain.ScoringRuleSet
	contestRuleSets map[uuid.UUID]*domain.ScoringRuleSet
	fallbacks       map[uuid.UUID]*domain.ScoringRuleSet
}

func (m *mockScorePreviewRepository) FetchOngoingContestRegistrations(context.Context, *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	return m.registrations, nil
}

func (m *mockScorePreviewRepository) FindUnitForTracking(_ context.Context, req *domain.UnitFindForTrackingRequest) (*domain.Unit, error) {
	return &domain.Unit{
		ID:            req.ID,
		Key:           domain.UnitKeyReadingPage,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (m *mockScorePreviewRepository) FindUnitForTrackingByKey(_ context.Context, req *domain.UnitFindForTrackingByKeyRequest) (*domain.Unit, error) {
	return &domain.Unit{
		ID:            uuid.New(),
		Key:           req.Key,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (m *mockScorePreviewRepository) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	return m.platformRuleSet, nil
}

func (m *mockScorePreviewRepository) FindContestScoringRuleSets(_ context.Context, contestID uuid.UUID) (*domain.ScoringRuleSet, *domain.ScoringRuleSet, error) {
	return m.contestRuleSets[contestID], m.fallbacks[contestID], nil
}

func TestScorePreviewExecute(t *testing.T) {
	userID := uuid.New()
	unitID := uuid.New()
	registrationID := uuid.New()
	contestID := uuid.New()
	platformRuleSetID := uuid.New()
	platformRuleID := uuid.New()
	contestRuleSetID := uuid.New()
	contestRuleID := uuid.New()
	amount := float32(100)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	repo := &mockScorePreviewRepository{
		registrations: &domain.ContestRegistrations{Registrations: []domain.ContestRegistration{{
			ID:        registrationID,
			ContestID: contestID,
			Languages: []domain.Language{{Code: "jpn"}},
			Contest: &domain.ContestView{AllowedActivities: []domain.Activity{{
				ID: 1,
			}}},
		}}},
		platformRuleSet: &domain.ScoringRuleSet{
			ID: platformRuleSetID,
			Rules: []domain.ScoringRule{{
				ID:          platformRuleID,
				Priority:    1,
				ActivityID:  1,
				UnitKey:     domain.UnitKeyReadingPage,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        2,
			}},
		},
		contestRuleSets: map[uuid.UUID]*domain.ScoringRuleSet{
			contestID: {
				ID:   contestRuleSetID,
				Mode: domain.ScoringRuleSetModeReplace,
				Rules: []domain.ScoringRule{{
					ID:           contestRuleID,
					Priority:     1,
					ActivityID:   1,
					UnitKey:      domain.UnitKeyReadingPage,
					LanguageCode: "jpn",
					ScoreSource:  domain.ScoreSourceAmount,
					Rate:         0.5,
				}},
			},
		},
	}
	service := domain.NewScorePreview(repo, commondomain.NewMockClock(now))

	result, err := service.Execute(ctxWithUserSubject(userID.String()), &domain.ScorePreviewRequest{
		RegistrationIDs: []uuid.UUID{registrationID},
		UnitID:          &unitID,
		ActivityID:      1,
		LanguageCode:    "jpn",
		Amount:          &amount,
		Tags:            []string{" Book "},
	})

	require.NoError(t, err)
	assert.Equal(t, float32(200), result.Platform.Score)
	assert.Equal(t, &platformRuleSetID, result.Platform.RuleSetID)
	assert.Equal(t, []domain.AppliedScoringRule{{RuleID: platformRuleID, Rate: 2}}, result.Platform.Rules)
	require.Len(t, result.Contests, 1)
	assert.Equal(t, registrationID, result.Contests[0].RegistrationID)
	assert.Equal(t, contestID, result.Contests[0].ContestID)
	assert.Equal(t, float32(50), result.Contests[0].Estimate.Score)
	assert.Equal(t, &contestRuleSetID, result.Contests[0].Estimate.RuleSetID)
	assert.Equal(t, []domain.AppliedScoringRule{{RuleID: contestRuleID, Rate: 0.5}}, result.Contests[0].Estimate.Rules)
}

func TestScorePreviewExecuteReturnsZeroForUncoveredInput(t *testing.T) {
	amount := float32(10)
	repo := &mockScorePreviewRepository{
		platformRuleSet: &domain.ScoringRuleSet{
			ID:    uuid.New(),
			Rules: []domain.ScoringRule{},
		},
	}
	service := domain.NewScorePreview(repo, commondomain.NewMockClock(time.Now()))

	result, err := service.Execute(ctxWithUserSubject(uuid.NewString()), &domain.ScorePreviewRequest{
		UnitKey:      ptr(domain.UnitKeyReadingPage),
		ActivityID:   1,
		LanguageCode: "jpn",
		Amount:       &amount,
	})

	require.NoError(t, err)
	assert.Zero(t, result.Platform.Score)
	assert.Nil(t, result.Platform.RuleSetID)
	assert.Empty(t, result.Platform.Rules)
}

func TestScorePreviewExecuteRequiresAuthentication(t *testing.T) {
	service := domain.NewScorePreview(&mockScorePreviewRepository{}, commondomain.NewMockClock(time.Now()))

	_, err := service.Execute(ctxWithGuest(), &domain.ScorePreviewRequest{})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func ptr[T any](value T) *T {
	return &value
}
