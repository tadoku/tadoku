package domain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLogContestUpdateRepository struct {
	registrations    *domain.ContestRegistrations
	fetchRegErr      error
	log              *domain.Log
	updatedLog       *domain.Log
	findErr          error
	updateErr        error
	updateCalled     bool
	updateCalledWith *domain.LogContestUpdateDBRequest
	findCallCount    int
}

func (m *mockLogContestUpdateRepository) FetchOngoingContestRegistrations(ctx context.Context, req *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	return m.registrations, m.fetchRegErr
}

func (m *mockLogContestUpdateRepository) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	m.findCallCount++
	if m.findCallCount > 1 && m.updatedLog != nil {
		return m.updatedLog, m.findErr
	}
	return m.log, m.findErr
}

func (m *mockLogContestUpdateRepository) UpdateLogContests(ctx context.Context, req *domain.LogContestUpdateDBRequest) error {
	m.updateCalled = true
	m.updateCalledWith = req
	return m.updateErr
}

func newLogContestUpdateService(repo *mockLogContestUpdateRepository, clock commondomain.Clock) *domain.LogContestUpdate {
	return domain.NewLogContestUpdate(repo, clock)
}

func TestLogContestUpdate_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	logID := uuid.New()
	registrationID := uuid.New()
	contestID := uuid.New()
	now := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)

	baseLog := &domain.Log{
		ID:           logID,
		UserID:       userID,
		LanguageCode: "jpn",
		ActivityID:   1,
		Amount:       100,
		Modifier:     1.0,
		Score:        100,
		CreatedAt:    now,
	}

	ongoingRegistrations := &domain.ContestRegistrations{
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
					ContestEnd: now.Add(30 * 24 * time.Hour),
				},
			},
		},
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockLogContestUpdateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithGuest()
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockLogContestUpdateRepository{}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		_, err := svc.Execute(context.Background(), &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns forbidden when user is not log owner", func(t *testing.T) {
		repo := &mockLogContestUpdateRepository{
			log: &domain.Log{
				ID:     logID,
				UserID: otherUserID,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error when log not found", func(t *testing.T) {
		repo := &mockLogContestUpdateRepository{
			findErr: domain.ErrNotFound,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.Error(t, err)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error when registration not in ongoing set", func(t *testing.T) {
		badRegID := uuid.New()
		repo := &mockLogContestUpdateRepository{
			log:           baseLog,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{badRegID},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error when language does not match registration", func(t *testing.T) {
		logWithKorean := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "kor",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			CreatedAt:    now,
		}
		repo := &mockLogContestUpdateRepository{
			log:           logWithKorean,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("returns error when activity does not match contest", func(t *testing.T) {
		logWithListening := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   999,
			Amount:       100,
			Modifier:     1.0,
			CreatedAt:    now,
		}
		repo := &mockLogContestUpdateRepository{
			log:           logWithListening,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("successfully attaches a registration", func(t *testing.T) {
		updatedLog := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
			Registrations: []domain.ContestRegistrationReference{
				{RegistrationID: registrationID, ContestID: contestID, ContestEnd: now.Add(30 * 24 * time.Hour), Title: "Test"},
			},
		}
		repo := &mockLogContestUpdateRepository{
			log:           baseLog, // no registrations initially
			updatedLog:    updatedLog,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		result, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		require.Len(t, repo.updateCalledWith.Attach, 1)
		assert.Equal(t, registrationID, repo.updateCalledWith.Attach[0].RegistrationID)
		assert.Equal(t, contestID, repo.updateCalledWith.Attach[0].ContestID)
		assert.Empty(t, repo.updateCalledWith.Detach)
		assert.Equal(t, logID, result.ID)
	})

	t.Run("successfully detaches a registration", func(t *testing.T) {
		logWithRegistration := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
			Registrations: []domain.ContestRegistrationReference{
				{RegistrationID: registrationID, ContestID: contestID, ContestEnd: now.Add(30 * 24 * time.Hour), Title: "Test"},
			},
		}
		updatedLog := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
		}
		repo := &mockLogContestUpdateRepository{
			log:           logWithRegistration,
			updatedLog:    updatedLog,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{}, // empty = detach all
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Empty(t, repo.updateCalledWith.Attach)
		require.Len(t, repo.updateCalledWith.Detach, 1)
		assert.Equal(t, contestID, repo.updateCalledWith.Detach[0])
	})

	t.Run("no-op when desired set matches current ongoing set", func(t *testing.T) {
		logWithRegistration := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
			Registrations: []domain.ContestRegistrationReference{
				{RegistrationID: registrationID, ContestID: contestID, ContestEnd: now.Add(30 * 24 * time.Hour), Title: "Test"},
			},
		}
		repo := &mockLogContestUpdateRepository{
			log:           logWithRegistration,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		result, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		require.NoError(t, err)
		assert.False(t, repo.updateCalled) // no changes needed
		assert.Equal(t, logID, result.ID)
	})

	t.Run("returns error when UpdateLogContests fails", func(t *testing.T) {
		repo := &mockLogContestUpdateRepository{
			log:           baseLog, // no registrations initially
			registrations: ongoingRegistrations,
			updateErr:     fmt.Errorf("database error"),
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{registrationID},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not update log contests")
		assert.True(t, repo.updateCalled)
	})

	t.Run("preserves ended contest attachments", func(t *testing.T) {
		endedContestID := uuid.New()
		endedRegID := uuid.New()
		logWithEndedAndOngoing := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
			Registrations: []domain.ContestRegistrationReference{
				{RegistrationID: registrationID, ContestID: contestID, ContestEnd: now.Add(30 * 24 * time.Hour), Title: "Ongoing"},
				{RegistrationID: endedRegID, ContestID: endedContestID, ContestEnd: now.Add(-48 * time.Hour), Title: "Ended"},
			},
		}
		updatedLog := &domain.Log{
			ID:           logID,
			UserID:       userID,
			LanguageCode: "jpn",
			ActivityID:   1,
			Amount:       100,
			Modifier:     1.0,
			Score:        100,
			CreatedAt:    now,
		}
		repo := &mockLogContestUpdateRepository{
			log:           logWithEndedAndOngoing,
			updatedLog:    updatedLog,
			registrations: ongoingRegistrations,
		}
		clock := commondomain.NewMockClock(now)
		svc := newLogContestUpdateService(repo, clock)

		ctx := ctxWithUserSubject(userID.String())
		// Send empty desired set — should only detach ongoing, not ended
		_, err := svc.Execute(ctx, &domain.LogContestUpdateRequest{
			LogID:           logID,
			RegistrationIDs: []uuid.UUID{},
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		// Only the ongoing contest should be detached
		require.Len(t, repo.updateCalledWith.Detach, 1)
		assert.Equal(t, contestID, repo.updateCalledWith.Detach[0])
		// Ended contest is NOT in the detach list
	})
}
