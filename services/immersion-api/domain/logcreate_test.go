package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLogCreateRepository struct {
	registrations     *domain.ContestRegistrations
	fetchRegErr       error
	unit              *domain.Unit
	findUnitErr       error
	scoringRuleSet    *domain.ScoringRuleSet
	scoringRuleErr    error
	scoringRuleCalls  int
	contestRuleSets   map[uuid.UUID]*domain.ScoringRuleSet
	contestFallbacks  map[uuid.UUID]*domain.ScoringRuleSet
	contestScoringErr error
	createdLogID      *uuid.UUID
	createErr         error
	log               *domain.Log
	findErr           error
	createCalled      bool
	createCalledWith  *domain.LogCreateRequest
}

func (m *mockLogCreateRepository) FindContestScoringRuleSets(_ context.Context, contestID uuid.UUID) (*domain.ScoringRuleSet, *domain.ScoringRuleSet, error) {
	if m.contestScoringErr != nil {
		return nil, nil, m.contestScoringErr
	}
	return m.contestRuleSets[contestID], m.contestFallbacks[contestID], nil
}

func (m *mockLogCreateRepository) FetchOngoingContestRegistrations(ctx context.Context, req *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	return m.registrations, m.fetchRegErr
}

func (m *mockLogCreateRepository) FindUnitForTracking(_ context.Context, req *domain.UnitFindForTrackingRequest) (*domain.Unit, error) {
	if m.findUnitErr != nil {
		return nil, m.findUnitErr
	}
	if m.unit != nil {
		return m.unit, nil
	}
	return &domain.Unit{
		ID:            req.ID,
		Key:           domain.UnitKeyReadingPage,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (m *mockLogCreateRepository) FindUnitForTrackingByKey(_ context.Context, req *domain.UnitFindForTrackingByKeyRequest) (*domain.Unit, error) {
	if m.findUnitErr != nil {
		return nil, m.findUnitErr
	}
	if m.unit != nil {
		return m.unit, nil
	}
	return &domain.Unit{
		ID:            uuid.New(),
		Key:           req.Key,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (m *mockLogCreateRepository) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	m.scoringRuleCalls++
	if m.scoringRuleErr != nil {
		return nil, m.scoringRuleErr
	}
	if m.scoringRuleSet != nil {
		return m.scoringRuleSet, nil
	}
	return defaultScoringShadowRuleSet(), nil
}

func defaultScoringShadowRuleSet() *domain.ScoringRuleSet {
	rules := make([]domain.ScoringRule, 0, 10)
	for activityID := int32(1); activityID <= 5; activityID++ {
		rules = append(rules,
			domain.ScoringRule{
				ID:          uuid.New(),
				Priority:    activityID * 10,
				ActivityID:  activityID,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        1,
			},
			domain.ScoringRule{
				ID:          uuid.New(),
				Priority:    activityID*10 + 1,
				ActivityID:  activityID,
				ScoreSource: domain.ScoreSourceDurationMinutes,
				Rate:        1,
			},
		)
	}
	return &domain.ScoringRuleSet{ID: uuid.New(), Rules: rules}
}

func (m *mockLogCreateRepository) CreateLog(ctx context.Context, req *domain.LogCreateRequest) (*uuid.UUID, error) {
	m.createCalled = true
	m.createCalledWith = req
	return m.createdLogID, m.createErr
}

func (m *mockLogCreateRepository) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	return m.log, m.findErr
}

type mockUserUpsertRepositoryForLog struct {
	err error
}

func (m *mockUserUpsertRepositoryForLog) UpsertUser(ctx context.Context, req *domain.UserUpsertRequest) error {
	return m.err
}

func newLogCreateService(repo *mockLogCreateRepository, clock commondomain.Clock) *domain.LogCreate {
	userRepo := &mockUserUpsertRepositoryForLog{}
	userUpsert := domain.NewUserUpsert(userRepo)
	return domain.NewLogCreate(repo, clock, userUpsert)
}

func newLogCreateServiceWithScoringEngine(repo *mockLogCreateRepository, clock commondomain.Clock) *domain.LogCreate {
	userRepo := &mockUserUpsertRepositoryForLog{}
	userUpsert := domain.NewUserUpsert(userRepo)
	return domain.NewLogCreateWithScoringEngine(repo, clock, userUpsert, true)
}

func TestLogCreate_Execute(t *testing.T) {
	userID := uuid.New()
	logID := uuid.New()
	registrationID := uuid.New()
	contestID := uuid.New()
	unitID := uuid.New()
	amount100 := float32(100)
	durationZero := int32(0)
	duration600 := int32(600)
	duration900 := int32(900)
	now := time.Now()

	validRegistrations := &domain.ContestRegistrations{
		Registrations: []domain.ContestRegistration{
			{
				ID:        registrationID,
				ContestID: contestID,
				UserID:    userID,
				Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
				Contest: &domain.ContestView{
					ID:       contestID,
					Official: false,
					AllowedActivities: []domain.Activity{
						{ID: 1, Name: "Reading"},
					},
				},
			},
		},
	}

	createdLog := &domain.Log{
		ID:           logID,
		UserID:       userID,
		LanguageCode: "jpn",
		ActivityID:   1,
		Amount:       100,
		Score:        50,
		CreatedAt:    now,
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		_, err := svc.Execute(context.Background(), &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request (missing required fields)", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			// Missing required fields
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request with only amount", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request with only unit", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request with non-positive duration", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			ActivityID:      1,
			LanguageCode:    "jpn",
			DurationSeconds: &durationZero,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when registration not found for user", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{Registrations: []domain.ContestRegistration{}},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when language not allowed by registration", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "kor", // Not in registration
			Amount:          &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for unknown activity", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   999,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when activity not allowed by contest", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      2, // Valid activity, but not allowed
			LanguageCode:    "jpn",
			Amount:          &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("successfully creates log", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          &amount100,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalled)
		assert.Equal(t, logID, result.ID)
		assert.Equal(t, userID, repo.createCalledWith.UserID())
		tracking := repo.createCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingAmountUnit, tracking.Kind)
		assert.Equal(t, unitID, tracking.UnitID)
		assert.Equal(t, amount100, tracking.Amount)
		assert.Equal(t, float32(1), tracking.Modifier)
		assert.InDelta(t, float32(100), tracking.ComputedScore, 0.0001)
	})

	t.Run("successfully creates log without registration IDs", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalled)
		assert.Equal(t, logID, result.ID)
		assert.Empty(t, repo.createCalledWith.RegistrationIDs)
		assert.False(t, repo.createCalledWith.EligibleOfficialLeaderboard())
	})

	t.Run("successfully creates an amount log from a stable unit key", func(t *testing.T) {
		resolvedUnitID := uuid.New()
		unitKey := domain.UnitKeyReadingCharacter
		repo := &mockLogCreateRepository{
			unit: &domain.Unit{
				ID:            resolvedUnitID,
				Key:           unitKey,
				LogActivityID: 1,
				Modifier:      0.0025,
			},
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitKey:      &unitKey,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		require.NoError(t, err)
		tracking := repo.createCalledWith.Tracking()
		assert.Equal(t, resolvedUnitID, tracking.UnitID)
		assert.Equal(t, unitKey, tracking.UnitKey)
		assert.InDelta(t, float32(0.25), tracking.ComputedScore, 0.0001)
	})

	t.Run("rejects conflicting legacy and stable unit identifiers", func(t *testing.T) {
		unitKey := domain.UnitKeyReadingCharacter
		repo := &mockLogCreateRepository{
			unit: &domain.Unit{
				ID:            unitID,
				Key:           domain.UnitKeyReadingPage,
				LogActivityID: 1,
				Modifier:      1,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			UnitKey:      &unitKey,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("rejects a legacy unit that has no code-owned key", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			unit: &domain.Unit{
				ID:            unitID,
				Key:           "unknown",
				LogActivityID: 1,
				Modifier:      1,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("successfully creates duration-only log", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			ActivityID:      2,
			LanguageCode:    "jpn",
			DurationSeconds: &duration600,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalled)
		assert.Equal(t, logID, result.ID)
		tracking := repo.createCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingDuration, tracking.Kind)
		assert.Equal(t, duration600, tracking.DurationSeconds)
		assert.InDelta(t, float32(4), tracking.ComputedScore, 0.0001)
	})

	t.Run("keeps writing the interim score when shadow evaluation fails", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			scoringRuleErr: errors.New("scoring unavailable"),
			createdLogID:   &logID,
			log:            createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, repo.scoringRuleCalls)
		assert.True(t, repo.createCalled)
		assert.Equal(t, float32(100), repo.createCalledWith.Tracking().ComputedScore)
	})

	t.Run("writes the engine score and provenance when enabled", func(t *testing.T) {
		ruleSetID := uuid.New()
		ruleID := uuid.New()
		repo := &mockLogCreateRepository{
			scoringRuleSet: &domain.ScoringRuleSet{
				ID: ruleSetID,
				Rules: []domain.ScoringRule{{
					ID:          ruleID,
					Priority:    1,
					ActivityID:  1,
					UnitKey:     domain.UnitKeyReadingPage,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        2,
				}},
			},
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateServiceWithScoringEngine(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		require.NoError(t, err)
		tracking := repo.createCalledWith.Tracking()
		assert.Equal(t, float32(200), tracking.ComputedScore)
		require.NotNil(t, tracking.ScoreProvenance)
		assert.Equal(t, &ruleSetID, tracking.ScoreProvenance.RuleSetID)
		assert.Equal(t, []uuid.UUID{ruleID}, tracking.ScoreProvenance.RuleIDs)
		assert.Equal(t, []float32{2}, tracking.ScoreProvenance.Rates)
		assert.Equal(t, domain.ScoreSourceAmount, tracking.ScoreProvenance.Source)
	})

	t.Run("writes zero for an unmatched input when enabled", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			scoringRuleSet: &domain.ScoringRuleSet{
				ID: uuid.New(),
				Rules: []domain.ScoringRule{{
					ID:          uuid.New(),
					Priority:    1,
					ActivityID:  3,
					ScoreSource: domain.ScoreSourceAmount,
					Rate:        1,
				}},
			},
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateServiceWithScoringEngine(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		require.NoError(t, err)
		tracking := repo.createCalledWith.Tracking()
		assert.Zero(t, tracking.ComputedScore)
		require.NotNil(t, tracking.ScoreProvenance)
		assert.Nil(t, tracking.ScoreProvenance.RuleSetID)
		assert.Equal(t, domain.ScoreSourceAmount, tracking.ScoreProvenance.Source)
	})

	t.Run("snapshots an independent contest score when enabled", func(t *testing.T) {
		platformRuleSetID := uuid.New()
		contestRuleSetID := uuid.New()
		contestRuleID := uuid.New()
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
			scoringRuleSet: &domain.ScoringRuleSet{
				ID: platformRuleSetID,
				Rules: []domain.ScoringRule{{
					ID:          uuid.New(),
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
						ID:          contestRuleID,
						Priority:    1,
						ActivityID:  1,
						UnitKey:     domain.UnitKeyReadingPage,
						ScoreSource: domain.ScoreSourceAmount,
						Rate:        0.5,
					}},
				},
			},
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateServiceWithScoringEngine(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          &amount100,
		})

		require.NoError(t, err)
		assert.Equal(t, float32(200), repo.createCalledWith.Tracking().ComputedScore)
		contestTrackings := repo.createCalledWith.ContestTrackings()
		require.Len(t, contestTrackings, 1)
		assert.Equal(t, registrationID, contestTrackings[0].RegistrationID)
		assert.Equal(t, contestID, contestTrackings[0].ContestID)
		assert.Equal(t, float32(50), contestTrackings[0].Tracking.ComputedScore)
		require.NotNil(t, contestTrackings[0].Tracking.ScoreProvenance)
		assert.Equal(t, &contestRuleSetID, contestTrackings[0].Tracking.ScoreProvenance.RuleSetID)
		assert.Equal(t, []uuid.UUID{contestRuleID}, contestTrackings[0].Tracking.ScoreProvenance.RuleIDs)
	})

	t.Run("rejects the write when authoritative scoring fails", func(t *testing.T) {
		repo := &mockLogCreateRepository{scoringRuleErr: errors.New("scoring unavailable")}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateServiceWithScoringEngine(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogCreateRequest{
			UnitID:       &unitID,
			ActivityID:   1,
			LanguageCode: "jpn",
			Amount:       &amount100,
		})

		assert.Error(t, err)
		assert.False(t, repo.createCalled)
	})

	t.Run("successfully creates amount log with duration metadata", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			createdLogID: &logID,
			log:          createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          &amount100,
			DurationSeconds: &duration900,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalled)
		tracking := repo.createCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingBoth, tracking.Kind)
		assert.Equal(t, unitID, tracking.UnitID)
		assert.Equal(t, amount100, tracking.Amount)
		assert.Equal(t, duration900, tracking.DurationSeconds)
		assert.Equal(t, float32(1), tracking.Modifier)
		assert.InDelta(t, float32(100), tracking.ComputedScore, 0.0001)
	})

	t.Run("sets EligibleOfficialLeaderboard for official contest", func(t *testing.T) {
		officialRegistrations := &domain.ContestRegistrations{
			Registrations: []domain.ContestRegistration{
				{
					ID:        registrationID,
					ContestID: contestID,
					UserID:    userID,
					Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
					Contest: &domain.ContestView{
						ID:       contestID,
						Official: true,
						AllowedActivities: []domain.Activity{
							{ID: 1, Name: "Reading"},
						},
					},
				},
			},
		}

		repo := &mockLogCreateRepository{
			registrations: officialRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          &unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          &amount100,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalledWith.EligibleOfficialLeaderboard())
	})
}
