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

// mockLeaderboardUpdater implements the per-service leaderboard updater interfaces for testing.
// Used by services that only need to verify the updater was called correctly,
// without testing the updater's internal logic.
type mockLeaderboardUpdater struct {
	updateContestCalls  []mockUpdateContestCall
	updateOfficialCalls []mockUpdateOfficialCall
}

type mockUpdateContestCall struct {
	ContestID uuid.UUID
	UserID    uuid.UUID
}

type mockUpdateOfficialCall struct {
	Year   int
	UserID uuid.UUID
}

func (m *mockLeaderboardUpdater) UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) {
	m.updateContestCalls = append(m.updateContestCalls, mockUpdateContestCall{ContestID: contestID, UserID: userID})
}

func (m *mockLeaderboardUpdater) UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID) {
	m.updateOfficialCalls = append(m.updateOfficialCalls, mockUpdateOfficialCall{Year: year, UserID: userID})
}

func newLogCreateService(repo *mockLogCreateRepository, clock commondomain.Clock, updater *mockLeaderboardUpdater) *domain.LogCreate {
	userRepo := &mockUserUpsertRepositoryForLog{}
	userUpsert := domain.NewUserUpsert(userRepo)
	return domain.NewLogCreate(repo, clock, userUpsert, updater)
}

func TestLogCreate_Execute(t *testing.T) {
	userID := uuid.New()
	logID := uuid.New()
	registrationID := uuid.New()
	contestID := uuid.New()
	unitID := uuid.New()
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
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		_, err := svc.Execute(context.Background(), &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request (missing required fields)", func(t *testing.T) {
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			// Missing required fields
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when registration not found for user", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{Registrations: []domain.ContestRegistration{}},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when language not allowed by registration", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "kor", // Not in registration
			Amount:          100,
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when activity not allowed by contest", func(t *testing.T) {
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      999, // Not allowed
			LanguageCode:    "jpn",
			Amount:          100,
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
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalled)
		assert.Equal(t, logID, result.ID)
		assert.Equal(t, userID, repo.createCalledWith.UserID)
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
		svc := newLogCreateService(repo, clock, &mockLeaderboardUpdater{})

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.True(t, repo.createCalledWith.EligibleOfficialLeaderboard)
	})
}

func TestLogCreate_LeaderboardUpdates(t *testing.T) {
	userID := uuid.New()
	logID := uuid.New()
	registrationID := uuid.New()
	contestID := uuid.New()
	unitID := uuid.New()
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	t.Run("updates user contest score for each registration", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{
					{
						ID:        registrationID,
						ContestID: contestID,
						UserID:    userID,
						Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
						Contest: &domain.ContestView{
							ID:                contestID,
							Official:          false,
							AllowedActivities: []domain.Activity{{ID: 1, Name: "Reading"}},
						},
					},
				},
			},
			createdLogID: &logID,
			log: &domain.Log{
				ID:        logID,
				UserID:    userID,
				Score:     42.5,
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, updater.updateContestCalls, 1)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		assert.Equal(t, userID, updater.updateContestCalls[0].UserID)
		assert.Empty(t, updater.updateOfficialCalls)
	})

	t.Run("updates official scores for official contests", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{
					{
						ID:        registrationID,
						ContestID: contestID,
						UserID:    userID,
						Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
						Contest: &domain.ContestView{
							ID:                contestID,
							Official:          true,
							AllowedActivities: []domain.Activity{{ID: 1, Name: "Reading"}},
						},
					},
				},
			},
			createdLogID: &logID,
			log: &domain.Log{
				ID:        logID,
				UserID:    userID,
				Score:     50,
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, updater.updateContestCalls, 1)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		require.Len(t, updater.updateOfficialCalls, 1)
		assert.Equal(t, 2026, updater.updateOfficialCalls[0].Year)
		assert.Equal(t, userID, updater.updateOfficialCalls[0].UserID)
	})

	t.Run("does not update official scores for unofficial contests", func(t *testing.T) {
		updater := &mockLeaderboardUpdater{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{
					{
						ID:        registrationID,
						ContestID: contestID,
						UserID:    userID,
						Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
						Contest: &domain.ContestView{
							ID:                contestID,
							Official:          false,
							AllowedActivities: []domain.Activity{{ID: 1, Name: "Reading"}},
						},
					},
				},
			},
			createdLogID: &logID,
			log: &domain.Log{
				ID:        logID,
				UserID:    userID,
				Score:     50,
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.Len(t, updater.updateContestCalls, 1)
		assert.Empty(t, updater.updateOfficialCalls)
	})

	t.Run("updates multiple contest leaderboards", func(t *testing.T) {
		contestID2 := uuid.New()
		registrationID2 := uuid.New()

		updater := &mockLeaderboardUpdater{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
				Registrations: []domain.ContestRegistration{
					{
						ID:        registrationID,
						ContestID: contestID,
						UserID:    userID,
						Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
						Contest: &domain.ContestView{
							ID:                contestID,
							Official:          false,
							AllowedActivities: []domain.Activity{{ID: 1, Name: "Reading"}},
						},
					},
					{
						ID:        registrationID2,
						ContestID: contestID2,
						UserID:    userID,
						Languages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
						Contest: &domain.ContestView{
							ID:                contestID2,
							Official:          false,
							AllowedActivities: []domain.Activity{{ID: 1, Name: "Reading"}},
						},
					},
				},
			},
			createdLogID: &logID,
			log: &domain.Log{
				ID:        logID,
				UserID:    userID,
				Score:     25,
				CreatedAt: now,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, updater)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID, registrationID2},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, updater.updateContestCalls, 2)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		assert.Equal(t, contestID2, updater.updateContestCalls[1].ContestID)
	})
}
