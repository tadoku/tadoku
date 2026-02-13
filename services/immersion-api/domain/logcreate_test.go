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

type mockLeaderboardStore struct {
	incrementContestCalls []leaderboardIncrementCall
	incrementYearlyCalls  []leaderboardIncrementCall
	incrementGlobalCalls  []leaderboardIncrementCall
	rebuildContestCalls   []leaderboardRebuildCall
	rebuildYearlyCalls    []leaderboardRebuildCall
	rebuildGlobalCalls    int

	// Control behavior
	incrementContestExists bool
	incrementYearlyExists  bool
	incrementGlobalExists  bool
	incrementContestErr    error
	incrementYearlyErr     error
	incrementGlobalErr     error
	rebuildContestErr      error
	rebuildYearlyErr       error
	rebuildGlobalErr       error
}

type leaderboardIncrementCall struct {
	ContestID uuid.UUID
	Year      int
	UserID    uuid.UUID
	Score     float64
}

type leaderboardRebuildCall struct {
	ContestID uuid.UUID
	Year      int
	Scores    []domain.LeaderboardScore
}

func (m *mockLeaderboardStore) IncrementContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error) {
	m.incrementContestCalls = append(m.incrementContestCalls, leaderboardIncrementCall{
		ContestID: contestID, UserID: userID, Score: score,
	})
	return m.incrementContestExists, m.incrementContestErr
}

func (m *mockLeaderboardStore) IncrementYearlyScore(ctx context.Context, year int, userID uuid.UUID, score float64) (bool, error) {
	m.incrementYearlyCalls = append(m.incrementYearlyCalls, leaderboardIncrementCall{
		Year: year, UserID: userID, Score: score,
	})
	return m.incrementYearlyExists, m.incrementYearlyErr
}

func (m *mockLeaderboardStore) IncrementGlobalScore(ctx context.Context, userID uuid.UUID, score float64) (bool, error) {
	m.incrementGlobalCalls = append(m.incrementGlobalCalls, leaderboardIncrementCall{
		UserID: userID, Score: score,
	})
	return m.incrementGlobalExists, m.incrementGlobalErr
}

func (m *mockLeaderboardStore) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []domain.LeaderboardScore) error {
	m.rebuildContestCalls = append(m.rebuildContestCalls, leaderboardRebuildCall{
		ContestID: contestID, Scores: scores,
	})
	return m.rebuildContestErr
}

func (m *mockLeaderboardStore) RebuildYearlyLeaderboard(ctx context.Context, year int, scores []domain.LeaderboardScore) error {
	m.rebuildYearlyCalls = append(m.rebuildYearlyCalls, leaderboardRebuildCall{
		Year: year, Scores: scores,
	})
	return m.rebuildYearlyErr
}

func (m *mockLeaderboardStore) RebuildGlobalLeaderboard(ctx context.Context, scores []domain.LeaderboardScore) error {
	m.rebuildGlobalCalls++
	return m.rebuildGlobalErr
}

type mockLeaderboardRebuildRepo struct {
	contestScores []domain.LeaderboardScore
	yearlyScores  []domain.LeaderboardScore
	globalScores  []domain.LeaderboardScore
	contestErr    error
	yearlyErr     error
	globalErr     error
}

func (m *mockLeaderboardRebuildRepo) FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]domain.LeaderboardScore, error) {
	return m.contestScores, m.contestErr
}

func (m *mockLeaderboardRebuildRepo) FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]domain.LeaderboardScore, error) {
	return m.yearlyScores, m.yearlyErr
}

func (m *mockLeaderboardRebuildRepo) FetchAllGlobalLeaderboardScores(ctx context.Context) ([]domain.LeaderboardScore, error) {
	return m.globalScores, m.globalErr
}

func newLogCreateService(repo *mockLogCreateRepository, clock commondomain.Clock, store *mockLeaderboardStore, rebuildRepo *mockLeaderboardRebuildRepo) *domain.LogCreate {
	userRepo := &mockUserUpsertRepositoryForLog{}
	userUpsert := domain.NewUserUpsert(userRepo)
	return domain.NewLogCreate(repo, clock, userUpsert, store, rebuildRepo)
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
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		_, err := svc.Execute(context.Background(), &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request (missing required fields)", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			// Missing required fields
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when registration not found for user", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{Registrations: []domain.ContestRegistration{}},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

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
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

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
		store := &mockLeaderboardStore{}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

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
		store := &mockLeaderboardStore{incrementContestExists: true}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

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

		store := &mockLeaderboardStore{
			incrementContestExists: true,
			incrementYearlyExists:  true,
			incrementGlobalExists:  true,
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: officialRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

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

	t.Run("increments contest leaderboard when it exists in store", func(t *testing.T) {
		store := &mockLeaderboardStore{incrementContestExists: true}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, store.incrementContestCalls, 1)
		assert.Equal(t, contestID, store.incrementContestCalls[0].ContestID)
		assert.Equal(t, userID, store.incrementContestCalls[0].UserID)
		assert.InDelta(t, 42.5, store.incrementContestCalls[0].Score, 0.01)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("rebuilds contest leaderboard when not in store", func(t *testing.T) {
		dbScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 42.5},
			{UserID: uuid.New(), Score: 100},
		}
		store := &mockLeaderboardStore{incrementContestExists: false}
		rebuildRepo := &mockLeaderboardRebuildRepo{contestScores: dbScores}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, store.rebuildContestCalls, 1)
		assert.Equal(t, contestID, store.rebuildContestCalls[0].ContestID)
		assert.Equal(t, dbScores, store.rebuildContestCalls[0].Scores)
	})

	t.Run("updates yearly and global leaderboards for official contests", func(t *testing.T) {
		store := &mockLeaderboardStore{
			incrementContestExists: true,
			incrementYearlyExists:  true,
			incrementGlobalExists:  true,
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)

		// Contest leaderboard updated
		require.Len(t, store.incrementContestCalls, 1)
		assert.Equal(t, contestID, store.incrementContestCalls[0].ContestID)

		// Yearly leaderboard updated
		require.Len(t, store.incrementYearlyCalls, 1)
		assert.Equal(t, 2026, store.incrementYearlyCalls[0].Year)
		assert.Equal(t, userID, store.incrementYearlyCalls[0].UserID)
		assert.InDelta(t, 50, store.incrementYearlyCalls[0].Score, 0.01)

		// Global leaderboard updated
		require.Len(t, store.incrementGlobalCalls, 1)
		assert.Equal(t, userID, store.incrementGlobalCalls[0].UserID)
		assert.InDelta(t, 50, store.incrementGlobalCalls[0].Score, 0.01)
	})

	t.Run("does not update yearly or global for unofficial contests", func(t *testing.T) {
		store := &mockLeaderboardStore{incrementContestExists: true}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.Len(t, store.incrementContestCalls, 1)
		assert.Empty(t, store.incrementYearlyCalls)
		assert.Empty(t, store.incrementGlobalCalls)
	})

	t.Run("updates multiple contest leaderboards", func(t *testing.T) {
		contestID2 := uuid.New()
		registrationID2 := uuid.New()

		store := &mockLeaderboardStore{incrementContestExists: true}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID, registrationID2},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, store.incrementContestCalls, 2)
		assert.Equal(t, contestID, store.incrementContestCalls[0].ContestID)
		assert.Equal(t, contestID2, store.incrementContestCalls[1].ContestID)
	})

	t.Run("leaderboard store errors do not fail log creation", func(t *testing.T) {
		store := &mockLeaderboardStore{
			incrementContestErr: errors.New("redis connection refused"),
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{}
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.Equal(t, logID, result.ID)
	})

	t.Run("rebuild errors do not fail log creation", func(t *testing.T) {
		store := &mockLeaderboardStore{
			incrementContestExists: false,
		}
		rebuildRepo := &mockLeaderboardRebuildRepo{
			contestErr: errors.New("database timeout"),
		}
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
		svc := newLogCreateService(repo, clock, store, rebuildRepo)

		ctx := ctxWithUserSubject(userID.String())
		result, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.Equal(t, logID, result.ID)
	})
}
