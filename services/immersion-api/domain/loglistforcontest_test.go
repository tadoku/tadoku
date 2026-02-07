package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type logListForContestRepositoryMock struct {
	response        *domain.LogListForContestResponse
	err             error
	capturedRequest *domain.LogListForContestRequest
}

func (m *logListForContestRepositoryMock) ListLogsForContest(ctx context.Context, req *domain.LogListForContestRequest) (*domain.LogListForContestResponse, error) {
	m.capturedRequest = req
	return m.response, m.err
}

func TestLogListForContest_Execute(t *testing.T) {
	tests := []struct {
		name             string
		role             commondomain.Role
		requestPageSize  int
		includeDeleted   bool
		repoResponse     *domain.LogListForContestResponse
		repoErr          error
		expectedErr      error
		expectedPageSize int
	}{
		{
			name:            "default page size is 50",
			role:            commondomain.RoleUser,
			requestPageSize: 0,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:            "page size capped at 100",
			role:            commondomain.RoleUser,
			requestPageSize: 500,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "custom page size is preserved",
			role:            commondomain.RoleUser,
			requestPageSize: 25,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 25,
		},
		{
			name:           "admin can include deleted logs",
			role:           commondomain.RoleAdmin,
			includeDeleted: true,
			repoResponse: &domain.LogListForContestResponse{
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
			name:            "contest not found error is propagated",
			role:            commondomain.RoleUser,
			requestPageSize: 10,
			repoErr:         domain.ErrNotFound,
			expectedErr:     domain.ErrNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token := &commondomain.UserIdentity{
				Role:        test.role,
				Subject:     uuid.New().String(),
				DisplayName: "TestUser",
			}
			ctx := ctxWithToken(token)

			repo := &logListForContestRepositoryMock{
				response: test.repoResponse,
				err:      test.repoErr,
			}
			service := domain.NewLogListForContest(repo)

			result, err := service.Execute(ctx, &domain.LogListForContestRequest{
				ContestID:      uuid.New(),
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
