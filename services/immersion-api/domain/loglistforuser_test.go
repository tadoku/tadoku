package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

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
		role             commondomain.Role
		requestPageSize  int
		includeDeleted   bool
		repoResponse     *domain.LogListForUserResponse
		repoErr          error
		expectedErr      error
		expectedPageSize int
	}{
		{
			name:            "default page size is 50",
			role:            commondomain.RoleUser,
			requestPageSize: 0,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:            "page size capped at 100",
			role:            commondomain.RoleUser,
			requestPageSize: 500,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "negative page size defaults to 100",
			role:            commondomain.RoleUser,
			requestPageSize: -5,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "custom page size is preserved",
			role:            commondomain.RoleUser,
			requestPageSize: 25,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 25,
		},
		{
			name:           "admin can include deleted logs",
			role:           commondomain.RoleAdmin,
			includeDeleted: true,
			repoResponse: &domain.LogListForUserResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:           "non-admin cannot include deleted logs",
			role:           commondomain.RoleUser,
			includeDeleted: true,
			expectedErr:    domain.ErrUnauthorized,
		},
		{
			name:            "repository error is propagated",
			role:            commondomain.RoleUser,
			requestPageSize: 10,
			repoErr:         errors.New("database error"),
			expectedErr:     errors.New("database error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := &commondomain.UserIdentity{
				Role:        test.role,
				Subject:     uuid.New().String(),
				DisplayName: "TestUser",
			}
			ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, token)

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
				assert.Error(t, err)
				if errors.Is(test.expectedErr, domain.ErrUnauthorized) {
					assert.ErrorIs(t, err, domain.ErrUnauthorized)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
		})
	}
}
