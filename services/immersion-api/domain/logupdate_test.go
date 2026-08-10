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

type mockLogUpdateRepository struct {
	log               *domain.Log
	updatedLog        *domain.Log
	unit              *domain.Unit
	findUnitErr       error
	scoringRuleSet    *domain.ScoringRuleSet
	scoringRuleErr    error
	scoringRuleCalls  int
	contestRuleSets   map[uuid.UUID]*domain.ScoringRuleSet
	contestFallbacks  map[uuid.UUID]*domain.ScoringRuleSet
	contestScoringErr error
	findErr           error
	updateErr         error
	updateCalled      bool
	updateCalledWith  *domain.LogUpdateRequest
	findCallCount     int
}

func (m *mockLogUpdateRepository) FindContestScoringRuleSets(_ context.Context, contestID uuid.UUID) (*domain.ScoringRuleSet, *domain.ScoringRuleSet, error) {
	if m.contestScoringErr != nil {
		return nil, nil, m.contestScoringErr
	}
	return m.contestRuleSets[contestID], m.contestFallbacks[contestID], nil
}

func (m *mockLogUpdateRepository) FindLogByID(_ context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	m.findCallCount++
	if m.findCallCount > 1 && m.updatedLog != nil {
		return m.updatedLog, m.findErr
	}
	return m.log, m.findErr
}

func (m *mockLogUpdateRepository) FindUnitForTracking(_ context.Context, req *domain.UnitFindForTrackingRequest) (*domain.Unit, error) {
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

func (m *mockLogUpdateRepository) FindUnitForTrackingByKey(_ context.Context, req *domain.UnitFindForTrackingByKeyRequest) (*domain.Unit, error) {
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

func (m *mockLogUpdateRepository) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	m.scoringRuleCalls++
	if m.scoringRuleErr != nil {
		return nil, m.scoringRuleErr
	}
	if m.scoringRuleSet != nil {
		return m.scoringRuleSet, nil
	}
	return defaultScoringShadowRuleSet(), nil
}

func (m *mockLogUpdateRepository) UpdateLog(_ context.Context, req *domain.LogUpdateRequest) error {
	m.updateCalled = true
	m.updateCalledWith = req
	return m.updateErr
}

func TestLogUpdate_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	logID := uuid.New()
	unitID := uuid.New()
	amount10 := float32(10)
	amount15 := float32(15)
	amount20 := float32(20)
	durationZero := int32(0)
	duration600 := int32(600)
	duration900 := int32(900)
	now := time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)

	makeLog := func(ownerID uuid.UUID) *domain.Log {
		return &domain.Log{ID: logID, UserID: ownerID, ActivityID: 1}
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockLogUpdateRepository{log: makeLog(userID)}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogUpdateRepository{log: makeLog(userID)}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		_, err := svc.Execute(context.Background(), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.updateCalled)
	})

	t.Run("allows owner to update their own log", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 20}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount20,
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, updatedLog, result)
	})

	t.Run("returns forbidden for non-owner non-admin", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log: makeLog(otherUserID),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.updateCalled)
	})

	t.Run("allows admin to update any log", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: otherUserID, ActivityID: 1, Amount: 15}
		repo := &mockLogUpdateRepository{
			log:        makeLog(otherUserID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithAdminSubject(uuid.New().String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount15,
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, updatedLog, result)
	})

	t.Run("returns error when log not found", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			findErr: domain.ErrNotFound,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.Error(t, err)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error for invalid request missing unit", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log: makeLog(userID),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			Amount: &amount10,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error for invalid request missing amount", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log: makeLog(userID),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error for non-positive duration", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log: makeLog(userID),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:           logID,
			DurationSeconds: &durationZero,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log:       makeLog(userID),
			updateErr: errors.New("database error"),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.Error(t, err)
		assert.True(t, repo.updateCalled)
	})

	t.Run("sets now from clock and userID from log owner", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 10}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		require.NoError(t, err)
		assert.Equal(t, now, repo.updateCalledWith.Now())
		assert.Equal(t, userID, repo.updateCalledWith.UserID())
		tracking := repo.updateCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingAmountUnit, tracking.Kind)
		assert.Equal(t, unitID, tracking.UnitID)
		assert.Equal(t, amount10, tracking.Amount)
		assert.Equal(t, float32(1), tracking.Modifier)
		assert.InDelta(t, float32(10), tracking.ComputedScore, 0.0001)
	})

	t.Run("allows duration-only update", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 2, DurationSeconds: &duration600}
		repo := &mockLogUpdateRepository{
			log:        &domain.Log{ID: logID, UserID: userID, ActivityID: 2},
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:           logID,
			DurationSeconds: &duration600,
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, updatedLog, result)
		tracking := repo.updateCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingDuration, tracking.Kind)
		assert.Equal(t, duration600, tracking.DurationSeconds)
		assert.InDelta(t, float32(4), tracking.ComputedScore, 0.0001)
	})

	t.Run("keeps writing the interim score when shadow evaluation fails", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 10}
		repo := &mockLogUpdateRepository{
			log:            makeLog(userID),
			updatedLog:     updatedLog,
			scoringRuleErr: errors.New("scoring unavailable"),
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, repo.scoringRuleCalls)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, float32(10), repo.updateCalledWith.Tracking().ComputedScore)
	})

	t.Run("writes the engine score and provenance when enabled", func(t *testing.T) {
		ruleSetID := uuid.New()
		ruleID := uuid.New()
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 10}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
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
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdateWithScoringEngine(repo, clock, true)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		require.NoError(t, err)
		tracking := repo.updateCalledWith.Tracking()
		assert.Equal(t, float32(20), tracking.ComputedScore)
		require.NotNil(t, tracking.ScoreProvenance)
		assert.Equal(t, &ruleSetID, tracking.ScoreProvenance.RuleSetID)
		assert.Equal(t, []uuid.UUID{ruleID}, tracking.ScoreProvenance.RuleIDs)
	})

	t.Run("recomputes ongoing contest scores independently when enabled", func(t *testing.T) {
		contestID := uuid.New()
		registrationID := uuid.New()
		contestRuleSetID := uuid.New()
		contestRuleID := uuid.New()
		log := makeLog(userID)
		log.LanguageCode = "jpn"
		log.Registrations = []domain.ContestRegistrationReference{{
			RegistrationID: registrationID,
			ContestID:      contestID,
			ContestEnd:     now.Add(time.Hour),
		}}
		repo := &mockLogUpdateRepository{
			log:        log,
			updatedLog: &domain.Log{ID: logID, UserID: userID, ActivityID: 1},
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
						Rate:        3,
					}},
				},
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdateWithScoringEngine(repo, clock, true)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		require.NoError(t, err)
		contestTrackings := repo.updateCalledWith.ContestTrackings()
		require.Len(t, contestTrackings, 1)
		assert.Equal(t, contestID, contestTrackings[0].ContestID)
		assert.Equal(t, float32(30), contestTrackings[0].Tracking.ComputedScore)
		require.NotNil(t, contestTrackings[0].Tracking.ScoreProvenance)
		assert.Equal(t, &contestRuleSetID, contestTrackings[0].Tracking.ScoreProvenance.RuleSetID)
	})

	t.Run("allows amount update with duration metadata", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 10, DurationSeconds: &duration900}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:           logID,
			UnitID:          &unitID,
			Amount:          &amount10,
			DurationSeconds: &duration900,
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, updatedLog, result)
		tracking := repo.updateCalledWith.Tracking()
		assert.Equal(t, domain.LogTrackingBoth, tracking.Kind)
		assert.Equal(t, unitID, tracking.UnitID)
		assert.Equal(t, amount10, tracking.Amount)
		assert.Equal(t, duration900, tracking.DurationSeconds)
		assert.Equal(t, float32(1), tracking.Modifier)
		assert.InDelta(t, float32(10), tracking.ComputedScore, 0.0001)
	})

	t.Run("normalizes tags", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, ActivityID: 1, Amount: 10}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
			Tags:   []string{"Book", " FICTION ", "book"},
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction"}, repo.updateCalledWith.Tags)
	})

	t.Run("observes one update comparison in shadow mode", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: &domain.Log{ID: logID, UserID: userID, ActivityID: 1},
		}
		clock := commondomain.NewMockClock(now)
		observer := &recordingScoringShadowObserver{}
		svc := domain.NewLogUpdateWithScoringObserver(repo, clock, false, observer)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, repo.scoringRuleCalls)
		require.Len(t, observer.observations, 1)
		assert.Equal(t, domain.ScoringShadowOperationUpdate, observer.observations[0].Operation)
		assert.Equal(t, domain.ScoringShadowModeShadow, observer.observations[0].Mode)
		assert.Equal(t, domain.ScoringShadowOutcomeMatch, observer.observations[0].Outcome)
	})

	t.Run("observes an authoritative update error before rejecting the write", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log:            makeLog(userID),
			scoringRuleErr: errors.New("scoring unavailable"),
		}
		clock := commondomain.NewMockClock(now)
		observer := &recordingScoringShadowObserver{}
		svc := domain.NewLogUpdateWithScoringObserver(repo, clock, true, observer)

		_, err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: &unitID,
			Amount: &amount10,
		})

		assert.Error(t, err)
		assert.False(t, repo.updateCalled)
		assert.Equal(t, 1, repo.scoringRuleCalls)
		require.Len(t, observer.observations, 1)
		assert.Equal(t, domain.ScoringShadowOperationUpdate, observer.observations[0].Operation)
		assert.Equal(t, domain.ScoringShadowModeAuthoritative, observer.observations[0].Mode)
		assert.Equal(t, domain.ScoringShadowOutcomeError, observer.observations[0].Outcome)
	})
}
