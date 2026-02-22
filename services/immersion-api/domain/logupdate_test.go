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
	log              *domain.Log
	updatedLog       *domain.Log
	findErr          error
	updateErr        error
	updateCalled     bool
	updateCalledWith *domain.LogUpdateRequest
	findCallCount    int
	activityOverride *domain.Activity
}

func (m *mockLogUpdateRepository) FindLogByID(_ context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	m.findCallCount++
	if m.findCallCount > 1 && m.updatedLog != nil {
		return m.updatedLog, m.findErr
	}
	return m.log, m.findErr
}

func (m *mockLogUpdateRepository) UpdateLog(_ context.Context, req *domain.LogUpdateRequest) error {
	m.updateCalled = true
	m.updateCalledWith = req
	return m.updateErr
}

func (m *mockLogUpdateRepository) FindActivityByID(ctx context.Context, id int32) (*domain.Activity, error) {
	if m.activityOverride != nil {
		return m.activityOverride, nil
	}
	return &domain.Activity{ID: id, Name: "test", InputType: "amount", TimeModifier: 0.3}, nil
}

func TestLogUpdate_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	logID := uuid.New()
	unitID := uuid.New()
	now := time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC)

	makeLog := func(ownerID uuid.UUID) *domain.Log {
		return &domain.Log{ID: logID, UserID: ownerID}
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockLogUpdateRepository{log: makeLog(userID)}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithGuest()

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
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
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
		})

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.updateCalled)
	})

	t.Run("allows owner to update their own log", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, Amount: ptrFloat32(20)}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(20),
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
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
		})

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.updateCalled)
	})

	t.Run("allows admin to update any log", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: otherUserID, Amount: ptrFloat32(15)}
		repo := &mockLogUpdateRepository{
			log:        makeLog(otherUserID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithAdminSubject(uuid.New().String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(15),
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
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
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
			Amount: ptrFloat32(10),
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
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
		})

		assert.Error(t, err)
		assert.True(t, repo.updateCalled)
	})

	t.Run("sets now from clock and userID from log owner", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, Amount: ptrFloat32(10)}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
		})

		require.NoError(t, err)
		assert.Equal(t, now, repo.updateCalledWith.Now())
		assert.Equal(t, userID, repo.updateCalledWith.UserID())
	})

	t.Run("allows updating with duration for time-based activity", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, DurationSeconds: ptrInt32(3600)}
		repo := &mockLogUpdateRepository{
			log:        &domain.Log{ID: logID, UserID: userID, ActivityID: 3},
			updatedLog: updatedLog,
			activityOverride: &domain.Activity{
				ID: 3, Name: "Listening", InputType: "time", TimeModifier: 0.3,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		result, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:           logID,
			DurationSeconds: ptrInt32(3600),
		})

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, updatedLog, result)
		assert.NotNil(t, repo.updateCalledWith.Activity())
		assert.Equal(t, "time", repo.updateCalledWith.Activity().InputType)
	})

	t.Run("rejects time-based activity update without duration", func(t *testing.T) {
		repo := &mockLogUpdateRepository{
			log: &domain.Log{ID: logID, UserID: userID, ActivityID: 3},
			activityOverride: &domain.Activity{
				ID: 3, Name: "Listening", InputType: "time", TimeModifier: 0.3,
			},
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
		})

		assert.ErrorIs(t, err, domain.ErrInvalidLog)
		assert.False(t, repo.updateCalled)
	})

	t.Run("normalizes tags", func(t *testing.T) {
		updatedLog := &domain.Log{ID: logID, UserID: userID, Amount: ptrFloat32(10)}
		repo := &mockLogUpdateRepository{
			log:        makeLog(userID),
			updatedLog: updatedLog,
		}
		clock := commondomain.NewMockClock(now)
		svc := domain.NewLogUpdate(repo, clock)

		ctx := ctxWithUserSubject(userID.String())

		_, err := svc.Execute(ctx, &domain.LogUpdateRequest{
			LogID:  logID,
			UnitID: ptrUUID(unitID),
			Amount: ptrFloat32(10),
			Tags:   []string{"Book", " FICTION ", "book"},
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"book", "fiction"}, repo.updateCalledWith.Tags)
	})
}
