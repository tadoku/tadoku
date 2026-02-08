package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

var errLogListForUserDatabase = errors.New("database error")

type logListForUserRepositoryMock struct {
	response        *domain.LogListForUserResponse
	err             error
	capturedRequest *domain.LogListForUserRequest
}

func (m *logListForUserRepositoryMock) ListLogsForUser(ctx context.Context, req *domain.LogListForUserRequest) (*domain.LogListForUserResponse, error) {
	m.capturedRequest = req
	return m.response, m.err
}

func TestLogListForUser_Execute(t *testing.T) {
	tests := []struct {
		name             string
		admin            bool
		requestPageSize  int
		includeDeleted   bool
		repoResponse     *domain.LogListForUserResponse
		repoErr          error
		expectedErr      error
		expectedPageSize int
	}{
		{
			name:            "default page size is 50",
			admin:           false,
			requestPageSize: 0,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:            "page size capped at 100",
			admin:           false,
			requestPageSize: 500,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "negative page size defaults to 100",
			admin:           false,
			requestPageSize: -5,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "custom page size is preserved",
			admin:           false,
			requestPageSize: 25,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 25,
		},
		{
			name:           "admin can include deleted logs",
			admin:          true,
			includeDeleted: true,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:           "non-admin cannot include deleted logs",
			admin:          false,
			includeDeleted: true,
			expectedErr:    domain.ErrUnauthorized,
		},
		{
			name:            "repository error is propagated",
			admin:           false,
			requestPageSize: 10,
			repoErr:         errLogListForUserDatabase,
			expectedErr:     errLogListForUserDatabase,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			subject := uuid.New().String()
			var ctx context.Context
			if test.admin {
				ctx = ctxWithAdminSubject(subject)
			} else {
				ctx = ctxWithUserSubject(subject)
			}

			repo := &logListForUserRepositoryMock{
				response: test.repoResponse,
				err:      test.repoErr,
			}
			service := domain.NewLogListForUser(repo)

			result, err := service.Execute(ctx, &domain.LogListForUserRequest{
				UserID:         uuid.New(),
				PageSize:       test.requestPageSize,
				IncludeDeleted: test.includeDeleted,
			})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
		})
	}
}
