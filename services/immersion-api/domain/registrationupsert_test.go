package domain_test

import (
	"context"
	"errors"
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

	detachCalled      bool
	detachCalledWith  *domain.DetachContestLogsForLanguagesRequest
	detachErr         error
	languageExists    map[string]bool
	languageExistsErr error
	languageBatches   [][]string
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

func (m *mockRegistrationUpsertRepository) LanguagesExist(_ context.Context, codes []string) (bool, error) {
	m.languageBatches = append(m.languageBatches, append([]string(nil), codes...))
	if m.languageExistsErr != nil {
		return false, m.languageExistsErr
	}
	if m.languageExists == nil {
		return true, nil
	}
	for _, code := range codes {
		if !m.languageExists[code] {
			return false, nil
		}
	}
	return true, nil
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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"},
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.Equal(t, userID, repo.upsertCalledWith.UserID())
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
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		ctx := ctxWithUserSubject(userID.String())

		err := svc.Execute(ctx, &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"},
		})

		require.NoError(t, err)
		assert.True(t, repo.upsertCalled)
		assert.Equal(t, existingRegID, repo.upsertCalledWith.ID())
	})

	t.Run("allows known language codes", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:        validContest,
			findRegErr:     domain.ErrNotFound,
			languageExists: map[string]bool{"jpn": true, "kor": true},
		}
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "kor"},
		})

		require.NoError(t, err)
		assert.Equal(t, [][]string{{"jpn", "kor"}}, repo.languageBatches)
		assert.True(t, repo.upsertCalled)
	})

	t.Run("rejects an unknown language code for an unrestricted contest", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:        validContest,
			findRegErr:     domain.ErrNotFound,
			languageExists: map[string]bool{"invalid": false},
		}
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"invalid"},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidContestRegistration)
		assert.Equal(t, [][]string{{"invalid"}}, repo.languageBatches)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("rejects a mixed list containing an unknown language code", func(t *testing.T) {
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:        validContest,
			findRegErr:     domain.ErrNotFound,
			languageExists: map[string]bool{"jpn": true, "invalid": false},
		}
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn", "invalid"},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidContestRegistration)
		assert.Equal(t, [][]string{{"jpn", "invalid"}}, repo.languageBatches)
		assert.False(t, repo.upsertCalled)
	})

	t.Run("returns a language lookup error without upserting a registration", func(t *testing.T) {
		lookupErr := errors.New("language repository unavailable")
		userRepo := &mockUserUpsertRepositoryForReg{}
		userUpsert := domain.NewUserUpsert(userRepo)
		repo := &mockRegistrationUpsertRepository{
			contest:           validContest,
			findRegErr:        domain.ErrNotFound,
			languageExistsErr: lookupErr,
		}
		svc := domain.NewRegistrationUpsert(repo, userUpsert)

		err := svc.Execute(ctxWithUserSubject(userID.String()), &domain.RegistrationUpsertRequest{
			ContestID:     contestID,
			LanguageCodes: []string{"jpn"},
		})

		assert.ErrorIs(t, err, lookupErr)
		assert.Equal(t, [][]string{{"jpn"}}, repo.languageBatches)
		assert.False(t, repo.upsertCalled)
	})
}
