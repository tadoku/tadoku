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

// mockLeaderboardStore implements domain.LeaderboardStore for testing.
type mockLeaderboardStore struct {
	updateContestCalls   []updateContestCall
	updateOfficialCalls  []updateOfficialCall
	rebuildContestCalls  []rebuildContestCall
	rebuildOfficialCalls []rebuildOfficialCall

	// Control behavior
	updateContestExists  bool
	updateOfficialYearly bool
	updateOfficialGlobal bool
	updateContestErr     error
	updateOfficialErr    error
	rebuildContestErr    error
	rebuildOfficialErr   error
}

type updateContestCall struct {
	ContestID uuid.UUID
	UserID    uuid.UUID
	Score     float64
}

type updateOfficialCall struct {
	Year        int
	UserID      uuid.UUID
	YearlyScore float64
	GlobalScore float64
}

type rebuildContestCall struct {
	ContestID uuid.UUID
	Scores    []domain.LeaderboardScore
}

type rebuildOfficialCall struct {
	Year         int
	YearlyScores []domain.LeaderboardScore
	GlobalScores []domain.LeaderboardScore
}

func (m *mockLeaderboardStore) UpdateContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error) {
	m.updateContestCalls = append(m.updateContestCalls, updateContestCall{
		ContestID: contestID, UserID: userID, Score: score,
	})
	return m.updateContestExists, m.updateContestErr
}

func (m *mockLeaderboardStore) UpdateOfficialScores(ctx context.Context, year int, userID uuid.UUID, yearlyScore float64, globalScore float64) (bool, bool, error) {
	m.updateOfficialCalls = append(m.updateOfficialCalls, updateOfficialCall{
		Year: year, UserID: userID, YearlyScore: yearlyScore, GlobalScore: globalScore,
	})
	return m.updateOfficialYearly, m.updateOfficialGlobal, m.updateOfficialErr
}

func (m *mockLeaderboardStore) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []domain.LeaderboardScore) error {
	m.rebuildContestCalls = append(m.rebuildContestCalls, rebuildContestCall{
		ContestID: contestID, Scores: scores,
	})
	return m.rebuildContestErr
}

func (m *mockLeaderboardStore) RebuildOfficialLeaderboards(ctx context.Context, year int, yearlyScores []domain.LeaderboardScore, globalScores []domain.LeaderboardScore) error {
	m.rebuildOfficialCalls = append(m.rebuildOfficialCalls, rebuildOfficialCall{
		Year: year, YearlyScores: yearlyScores, GlobalScores: globalScores,
	})
	return m.rebuildOfficialErr
}

// mockLeaderboardRepo implements domain.LeaderboardRepository for testing.
type mockLeaderboardRepo struct {
	contestScores    []domain.LeaderboardScore
	yearlyScores     []domain.LeaderboardScore
	globalScores     []domain.LeaderboardScore
	userContestScore float64
	userYearlyScore  float64
	userGlobalScore  float64
	contestErr       error
	yearlyErr        error
	globalErr        error
	userContestErr   error
	userYearlyErr    error
	userGlobalErr    error
}

func (m *mockLeaderboardRepo) FetchAllContestLeaderboardScores(ctx context.Context, contestID uuid.UUID) ([]domain.LeaderboardScore, error) {
	return m.contestScores, m.contestErr
}

func (m *mockLeaderboardRepo) FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]domain.LeaderboardScore, error) {
	return m.yearlyScores, m.yearlyErr
}

func (m *mockLeaderboardRepo) FetchAllGlobalLeaderboardScores(ctx context.Context) ([]domain.LeaderboardScore, error) {
	return m.globalScores, m.globalErr
}

func (m *mockLeaderboardRepo) FetchUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) (float64, error) {
	return m.userContestScore, m.userContestErr
}

func (m *mockLeaderboardRepo) FetchUserYearlyScore(ctx context.Context, year int, userID uuid.UUID) (float64, error) {
	return m.userYearlyScore, m.userYearlyErr
}

func (m *mockLeaderboardRepo) FetchUserGlobalScore(ctx context.Context, userID uuid.UUID) (float64, error) {
	return m.userGlobalScore, m.userGlobalErr
}

// mockLeaderboardScoreUpdater implements domain.LeaderboardScoreUpdater for testing.
// Used by services that only need to verify the updater was called correctly,
// without testing the updater's internal logic.
type mockLeaderboardScoreUpdater struct {
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

func (m *mockLeaderboardScoreUpdater) UpdateUserContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID) {
	m.updateContestCalls = append(m.updateContestCalls, mockUpdateContestCall{ContestID: contestID, UserID: userID})
}

func (m *mockLeaderboardScoreUpdater) UpdateUserOfficialScores(ctx context.Context, year int, userID uuid.UUID) {
	m.updateOfficialCalls = append(m.updateOfficialCalls, mockUpdateOfficialCall{Year: year, UserID: userID})
}

func newLogCreateService(repo *mockLogCreateRepository, clock commondomain.Clock, store *mockLeaderboardStore, lbRepo *mockLeaderboardRepo) *domain.LogCreate {
	userRepo := &mockUserUpsertRepositoryForLog{}
	userUpsert := domain.NewUserUpsert(userRepo)
	updater := domain.NewLeaderboardUpdater(store, lbRepo)
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
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

		_, err := svc.Execute(context.Background(), &domain.LogCreateRequest{})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error for invalid request (missing required fields)", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			// Missing required fields
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when registration not found for user", func(t *testing.T) {
		store := &mockLeaderboardStore{}
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{
			registrations: &domain.ContestRegistrations{Registrations: []domain.ContestRegistration{}},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
		lbRepo := &mockLeaderboardRepo{}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
		store := &mockLeaderboardStore{updateContestExists: true}
		lbRepo := &mockLeaderboardRepo{userContestScore: 50}
		repo := &mockLogCreateRepository{
			registrations: validRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
			updateContestExists:  true,
			updateOfficialYearly: true,
			updateOfficialGlobal: true,
		}
		lbRepo := &mockLeaderboardRepo{userContestScore: 50, userYearlyScore: 50, userGlobalScore: 50}
		repo := &mockLogCreateRepository{
			registrations: officialRegistrations,
			createdLogID:  &logID,
			log:           createdLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogCreateService(repo, clock, store, lbRepo)

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

	t.Run("recalculates user contest score when leaderboard exists in store", func(t *testing.T) {
		store := &mockLeaderboardStore{updateContestExists: true}
		lbRepo := &mockLeaderboardRepo{userContestScore: 142.5}
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, store.updateContestCalls, 1)
		assert.Equal(t, contestID, store.updateContestCalls[0].ContestID)
		assert.Equal(t, userID, store.updateContestCalls[0].UserID)
		assert.InDelta(t, 142.5, store.updateContestCalls[0].Score, 0.01)
		assert.Empty(t, store.rebuildContestCalls)
	})

	t.Run("rebuilds contest leaderboard when not in store", func(t *testing.T) {
		dbScores := []domain.LeaderboardScore{
			{UserID: userID, Score: 42.5},
			{UserID: uuid.New(), Score: 100},
		}
		store := &mockLeaderboardStore{updateContestExists: false}
		lbRepo := &mockLeaderboardRepo{userContestScore: 42.5, contestScores: dbScores}
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
			updateContestExists:  true,
			updateOfficialYearly: true,
			updateOfficialGlobal: true,
		}
		lbRepo := &mockLeaderboardRepo{
			userContestScore: 50,
			userYearlyScore:  200,
			userGlobalScore:  500,
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)

		// Contest leaderboard updated with recalculated score
		require.Len(t, store.updateContestCalls, 1)
		assert.Equal(t, contestID, store.updateContestCalls[0].ContestID)

		// Official leaderboards updated with recalculated scores (pipelined)
		require.Len(t, store.updateOfficialCalls, 1)
		assert.Equal(t, 2026, store.updateOfficialCalls[0].Year)
		assert.Equal(t, userID, store.updateOfficialCalls[0].UserID)
		assert.InDelta(t, 200, store.updateOfficialCalls[0].YearlyScore, 0.01)
		assert.InDelta(t, 500, store.updateOfficialCalls[0].GlobalScore, 0.01)
	})

	t.Run("does not update yearly or global for unofficial contests", func(t *testing.T) {
		store := &mockLeaderboardStore{updateContestExists: true}
		lbRepo := &mockLeaderboardRepo{userContestScore: 50}
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		assert.Len(t, store.updateContestCalls, 1)
		assert.Empty(t, store.updateOfficialCalls)
	})

	t.Run("updates multiple contest leaderboards", func(t *testing.T) {
		contestID2 := uuid.New()
		registrationID2 := uuid.New()

		store := &mockLeaderboardStore{updateContestExists: true}
		lbRepo := &mockLeaderboardRepo{userContestScore: 25}
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogCreateRequest{
			RegistrationIDs: []uuid.UUID{registrationID, registrationID2},
			UnitID:          unitID,
			ActivityID:      1,
			LanguageCode:    "jpn",
			Amount:          100,
		})

		require.NoError(t, err)
		require.Len(t, store.updateContestCalls, 2)
		assert.Equal(t, contestID, store.updateContestCalls[0].ContestID)
		assert.Equal(t, contestID2, store.updateContestCalls[1].ContestID)
	})

	t.Run("leaderboard store errors do not fail log creation", func(t *testing.T) {
		store := &mockLeaderboardStore{
			updateContestErr: errors.New("redis connection refused"),
		}
		lbRepo := &mockLeaderboardRepo{userContestScore: 50}
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
			updateContestExists: false,
		}
		lbRepo := &mockLeaderboardRepo{
			userContestScore: 50,
			contestErr:       errors.New("database timeout"),
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
		svc := newLogCreateService(repo, clock, store, lbRepo)

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
