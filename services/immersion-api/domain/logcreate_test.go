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

type mockLogCreateRepository struct {
	registrations    *domain.ContestRegistrations
	fetchRegErr      error
	unit             *domain.Unit
	findUnitErr      error
	createdLogID     *uuid.UUID
	createErr        error
	log              *domain.Log
	findErr          error
	createCalled     bool
	createCalledWith *domain.LogCreateRequest
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
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
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
