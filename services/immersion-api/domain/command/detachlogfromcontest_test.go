package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/domain"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
)

type DetachLogFromContestRepositoryMock struct {
	command.Repository
	detachCalled  bool
	detachErr     error
	contest       *immersiondomain.ContestView
	contestErr    error
	log           *immersiondomain.Log
	logErr        error
	detachUserID  uuid.UUID
	detachRequest *command.DetachLogFromContestRequest
}

func (r *DetachLogFromContestRepositoryMock) FindContestByID(ctx context.Context, req *immersiondomain.ContestFindRequest) (*immersiondomain.ContestView, error) {
	return r.contest, r.contestErr
}

func (r *DetachLogFromContestRepositoryMock) FindLogByID(ctx context.Context, req *immersiondomain.LogFindRequest) (*immersiondomain.Log, error) {
	return r.log, r.logErr
}

func (r *DetachLogFromContestRepositoryMock) DetachLogFromContest(ctx context.Context, req *command.DetachLogFromContestRequest, userID uuid.UUID) error {
	r.detachCalled = true
	r.detachRequest = req
	r.detachUserID = userID
	return r.detachErr
}

func (r *DetachLogFromContestRepositoryMock) UpsertUser(ctx context.Context, req *command.UpsertUserRequest) error {
	return nil
}

func TestDetachLogFromContest(t *testing.T) {
	clock := domain.NewMockClock(time.Now())
	contestOwnerID := uuid.New()
	otherUserID := uuid.New()
	contestID := uuid.New()
	logID := uuid.New()

	validContest := &immersiondomain.ContestView{
		ID:          contestID,
		OwnerUserID: contestOwnerID,
		Title:       "Test Contest",
	}

	validLog := &immersiondomain.Log{
		ID:     logID,
		UserID: otherUserID,
	}

	tests := []struct {
		name         string
		request      *command.DetachLogFromContestRequest
		userID       uuid.UUID
		role         domain.Role
		contest      *immersiondomain.ContestView
		contestErr   error
		log          *immersiondomain.Log
		logErr       error
		detachErr    error
		expectedErr  error
		shouldDetach bool
	}{
		{
			name: "happy path - contest owner detaches log",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Spam content",
			},
			userID:       contestOwnerID,
			role:         domain.RoleUser,
			contest:      validContest,
			log:          validLog,
			expectedErr:  nil,
			shouldDetach: true,
		},
		{
			name: "happy path - admin detaches log from any contest",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Violates guidelines",
			},
			userID:       otherUserID,
			role:         domain.RoleAdmin,
			contest:      validContest,
			log:          validLog,
			expectedErr:  nil,
			shouldDetach: true,
		},
		{
			name: "guest user cannot detach logs",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       otherUserID,
			role:         domain.RoleGuest,
			expectedErr:  command.ErrUnauthorized,
			shouldDetach: false,
		},
		{
			name: "banned user cannot detach logs",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       otherUserID,
			role:         domain.RoleBanned,
			expectedErr:  command.ErrForbidden,
			shouldDetach: false,
		},
		{
			name: "non-owner non-admin cannot detach logs",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       otherUserID,
			role:         domain.RoleUser,
			contest:      validContest,
			log:          validLog,
			expectedErr:  command.ErrForbidden,
			shouldDetach: false,
		},
		{
			name: "contest not found",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       contestOwnerID,
			role:         domain.RoleUser,
			contestErr:   immersiondomain.ErrNotFound,
			expectedErr:  immersiondomain.ErrNotFound,
			shouldDetach: false,
		},
		{
			name: "log not found",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       contestOwnerID,
			role:         domain.RoleUser,
			contest:      validContest,
			logErr:       immersiondomain.ErrNotFound,
			expectedErr:  immersiondomain.ErrNotFound,
			shouldDetach: false,
		},
		{
			name: "detach operation fails",
			request: &command.DetachLogFromContestRequest{
				ContestID: contestID,
				LogID:     logID,
				Reason:    "Test",
			},
			userID:       contestOwnerID,
			role:         domain.RoleUser,
			contest:      validContest,
			log:          validLog,
			detachErr:    errors.New("database error"),
			expectedErr:  errors.New("database error"),
			shouldDetach: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := &domain.SessionToken{
				Role:        test.role,
				Subject:     test.userID.String(),
				DisplayName: "TestUser",
			}
			ctx := context.WithValue(context.Background(), domain.CtxSessionKey, token)

			repo := &DetachLogFromContestRepositoryMock{
				contest:    test.contest,
				contestErr: test.contestErr,
				log:        test.log,
				logErr:     test.logErr,
				detachErr:  test.detachErr,
			}
			service := command.NewService(repo, clock)

			err := service.DetachLogFromContest(ctx, test.request)

			// Check if detach was called as expected
			assert.Equal(t, test.shouldDetach, repo.detachCalled, "DetachLogFromContest should be called: %v, was called: %v", test.shouldDetach, repo.detachCalled)

			// Check error
			if test.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(test.expectedErr, command.ErrUnauthorized) ||
					errors.Is(test.expectedErr, command.ErrForbidden) ||
					errors.Is(test.expectedErr, immersiondomain.ErrNotFound) {
					assert.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			// If detach was called, verify the parameters
			if repo.detachCalled {
				assert.Equal(t, test.request, repo.detachRequest, "Request should be passed correctly")
				assert.Equal(t, test.userID, repo.detachUserID, "User ID should be passed correctly")
			}
		})
	}
}

func TestDetachLogFromContest_NoSession(t *testing.T) {
	clock := domain.NewMockClock(time.Now())
	ctx := context.Background() // No session in context

	repo := &DetachLogFromContestRepositoryMock{}
	service := command.NewService(repo, clock)

	err := service.DetachLogFromContest(ctx, &command.DetachLogFromContestRequest{
		ContestID: uuid.New(),
		LogID:     uuid.New(),
		Reason:    "Test",
	})

	assert.ErrorIs(t, err, command.ErrUnauthorized)
	assert.False(t, repo.detachCalled)
}
