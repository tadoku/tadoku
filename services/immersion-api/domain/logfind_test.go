package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

var errLogFindDatabase = errors.New("database error")

type logFindRepositoryMock struct {
	log             *domain.Log
	err             error
	capturedRequest *domain.LogFindRequest
}

func (m *logFindRepositoryMock) FindLogByID(ctx context.Context, req *domain.LogFindRequest) (*domain.Log, error) {
	m.capturedRequest = req
	return m.log, m.err
}

func TestLogFind_Execute(t *testing.T) {
	logID := uuid.New()
	ownerUserID := uuid.New()
	otherUserID := uuid.New()

	baseLog := &domain.Log{
		ID:           logID,
		UserID:       ownerUserID,
		LanguageCode: "jpn",
		LanguageName: "Japanese",
		ActivityID:   1,
		ActivityName: "Reading",
		UnitName:     ptrString("Pages"),
		Amount:       ptrFloat32(100),
		Modifier:     ptrFloat32(1.0),
		Score:        100,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Registrations: []domain.ContestRegistrationReference{
			{
				RegistrationID: uuid.New(),
				ContestID:      uuid.New(),
				ContestEnd:     time.Now().Add(24 * time.Hour),
				Title:          "Test Contest",
			},
		},
	}

	tests := []struct {
		name                   string
		userID                 uuid.UUID
		admin                  bool
		repoLog                *domain.Log
		repoErr                error
		expectedErr            error
		expectIncludeDeleted   bool
		expectRegistrationsNil bool
	}{
		{
			name:                   "owner can see registrations",
			userID:                 ownerUserID,
			repoLog:                copyLog(baseLog),
			expectedErr:            nil,
			expectIncludeDeleted:   false,
			expectRegistrationsNil: false,
		},
		{
			name:                   "admin can see registrations and deleted logs",
			userID:                 otherUserID,
			admin:                  true,
			repoLog:                copyLog(baseLog),
			expectedErr:            nil,
			expectIncludeDeleted:   true,
			expectRegistrationsNil: false,
		},
		{
			name:                   "non-owner cannot see registrations",
			userID:                 otherUserID,
			repoLog:                copyLog(baseLog),
			expectedErr:            nil,
			expectIncludeDeleted:   false,
			expectRegistrationsNil: true,
		},
		{
			name:        "log not found",
			userID:      ownerUserID,
			repoErr:     domain.ErrNotFound,
			expectedErr: domain.ErrNotFound,
		},
		{
			name:        "repository error",
			userID:      ownerUserID,
			repoErr:     errLogFindDatabase,
			expectedErr: errLogFindDatabase,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ctx context.Context
			if test.admin {
				ctx = ctxWithAdminSubject(test.userID.String())
			} else {
				ctx = ctxWithUserSubject(test.userID.String())
			}

			repo := &logFindRepositoryMock{
				log: test.repoLog,
				err: test.repoErr,
			}
			service := domain.NewLogFind(repo)

			result, err := service.Execute(ctx, &domain.LogFindRequest{ID: logID})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectIncludeDeleted, repo.capturedRequest.IncludeDeleted)

			if test.expectRegistrationsNil {
				assert.Nil(t, result.Registrations)
			} else {
				assert.NotNil(t, result.Registrations)
				assert.Len(t, result.Registrations, 1)
			}
		})
	}
}

func TestLogFind_NoSession(t *testing.T) {
	logID := uuid.New()
	ownerUserID := uuid.New()

	repo := &logFindRepositoryMock{
		log: &domain.Log{
			ID:     logID,
			UserID: ownerUserID,
		},
	}
	service := domain.NewLogFind(repo)

	ctx := context.Background() // No session

	result, err := service.Execute(ctx, &domain.LogFindRequest{ID: logID})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Nil(t, result)
}

func copyLog(l *domain.Log) *domain.Log {
	copy := *l
	copy.Registrations = make([]domain.ContestRegistrationReference, len(l.Registrations))
	for i, r := range l.Registrations {
		copy.Registrations[i] = r
	}
	return &copy
}
