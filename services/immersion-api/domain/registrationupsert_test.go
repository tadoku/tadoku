package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockRegistrationUpsertRepository struct {
	contest          *domain.ContestView
	findContestErr   error
	registration     *domain.ContestRegistration
	findRegErr       error
	upsertErr        error
	upsertCalled     bool
	upsertCalledWith *domain.RegistrationUpsertRequest

	detachCalled     bool
	detachCalledWith *domain.DetachContestLogsForLanguagesRequest
	detachErr        error
}

func (m *mockRegistrationUpsertRepository) FindContestByID(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
	return m.contest, m.findContestErr
}

func (m *mockRegistrationUpsertRepository) FindRegistrationForUser(ctx context.Context, req *domain.RegistrationFindRequest) (*domain.ContestRegistration, error) {
	return m.registration, m.findRegErr
}

func (m *mockRegistrationUpsertRepository) UpsertContestRegistration(ctx context.Context, req *domain.RegistrationUpsertRequest) error {
	m.upsertCalled = true
	m.upsertCalledWith = req
	return m.upsertErr
}

func (m *mockRegistrationUpsertRepository) DetachContestLogsForLanguages(ctx context.Context, req *domain.DetachContestLogsForLanguagesRequest) error {
	m.detachCalled = true
	m.detachCalledWith = req
	return m.detachErr
}

type mockUserUpsertRepositoryForReg struct {
	err error
}

func (m *mockUserUpsertRepositoryForReg) UpsertUser(ctx context.Context, req *domain.UserUpsertRequest) error {
	return m.err
}

func TestRegistrationUpsert_Execute(t *testing.T) {
	userID := uuid.New()
	contestID := uuid.New()
	now := time.Now()

	validContest := &domain.ContestView{
		ID:               contestID,
		ContestStart:     now.Add(-time.Hour),
		ContestEnd:       now.Add(time.Hour * 24),
		RegistrationEnd:  now.Add(time.Hour * 12),
		Title:            "Test Contest",
		OwnerUserID:      uuid.New(),
		AllowedLanguages: []domain.Language{},
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithGuest()

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn"},
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		err := svc.Execute(context.Background(), &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn"},
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("returns error for invalid language count (zero)", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:    validContest,
			findRegErr: domain.ErrNotFound,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidContestRegistration)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("returns error for invalid language count (more than 3)", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:    validContest,
			findRegErr: domain.ErrNotFound,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor", "zho", "eng"},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidContestRegistration)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("returns error when language not allowed by contest", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		contestWithAllowList := &domain.ContestView{
			ID:               contestID,
			ContestStart:     now.Add(-time.Hour),
			ContestEnd:       now.Add(time.Hour * 24),
			RegistrationEnd:  now.Add(time.Hour * 12),
			Title:            "Test Contest",
			OwnerUserID:      uuid.New(),
			AllowedLanguages: []domain.Language{{Code: "jpn", Name: "Japanese"}},
		}
		repo := &mockRegistrationUpsertRepository{
			contest:    contestWithAllowList,
			findRegErr: domain.ErrNotFound,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"kor"},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidContestRegistration)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("detaches logs when removing a previously registered language", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		existingRegID := uuid.New()
		existingRegistration := &domain.ContestRegistration{
			ID:        existingRegID,
			ContestID: contestID,
			UserID:    userID,
			Languages: []domain.Language{
				{Code: "jpn", Name: "Japanese"},
				{Code: "kor", Name: "Korean"},
			},
		}
		repo := &mockRegistrationUpsertRepository{
			contest:      validContest,
			registration: existingRegistration,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn"}, // Removing "kor"
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.True(t, repo.detachCalled)
		assert.Equal(t, contestID, repo.detachCalledWith.ContestID)
		assert.Equal(t, []string{"kor"}, repo.detachCalledWith.LanguageCodes)
	})

	t.Run("does not detach logs when no languages are removed", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		existingRegID := uuid.New()
		existingRegistration := &domain.ContestRegistration{
			ID:        existingRegID,
			ContestID: contestID,
			UserID:    userID,
			Languages: []domain.Language{
				{Code: "jpn", Name: "Japanese"},
			},
		}
		repo := &mockRegistrationUpsertRepository{
			contest:      validContest,
			registration: existingRegistration,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"}, // Adding "kor", keeping "jpn"
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.False(t, repo.detachCalled)
	})

	t.Run("successfully creates new registration", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:    validContest,
			findRegErr: domain.ErrNotFound,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"},
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.Equal(t, userID, repo.upsertCalledWith.UserID)
		assert.Equal(t, contestID, repo.upsertCalledWith.ContestID)
	})

	t.Run("successfully updates existing registration with additional language", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		existingRegID := uuid.New()
		existingRegistration := &domain.ContestRegistration{
			ID:        existingRegID,
			ContestID: contestID,
			UserID:    userID,
			Languages: []domain.Language{
				{Code: "jpn", Name: "Japanese"},
			},
		}
		repo := &mockRegistrationUpsertRepository{
			contest:      validContest,
			registration: existingRegistration,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"},
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.Equal(t, existingRegID, repo.upsertCalledWith.ID)
	})

	t.Run("updates user contest score after successful registration", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:    validContest,
			findRegErr: domain.ErrNotFound,
		}
		updater := &mockLeaderboardUpdater{}
		svc := domain.NewRegistrationUpsert(repo, userUpsert, updater)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn"},
		})

		require.NoError(t, err)
		require.Len(t, updater.updateContestCalls, 1)
		assert.Equal(t, contestID, updater.updateContestCalls[0].ContestID)
		assert.Equal(t, userID, updater.updateContestCalls[0].UserID)
	})
}
