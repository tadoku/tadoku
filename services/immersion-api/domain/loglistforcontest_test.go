package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
		admin            bool
		requestPageSize  int
		includeDeleted   bool
		repoResponse     *domain.LogListForContestResponse
		repoErr          error
		expectedErr      error
		expectedPageSize int
	}{
		{
			name:            "default page size is 50",
			admin:           false,
			requestPageSize: 0,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 50,
		},
		{
			name:            "page size capped at 100",
			admin:           false,
			requestPageSize: 500,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 100,
		},
		{
			name:            "custom page size is preserved",
			admin:           false,
			requestPageSize: 25,
			repoResponse: &domain.LogListForContestResponse{
				Logs:      []domain.Log{},
				TotalSize: 0,
			},
			expectedPageSize: 25,
		},
		{
			name:           "admin can include deleted logs",
			admin:          true,
			includeDeleted: true,
			repoResponse: &domain.LogListForContestResponse{
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
			name:            "contest not found error is propagated",
			admin:           false,
			requestPageSize: 10,
			repoErr:         domain.ErrNotFound,
			expectedErr:     domain.ErrNotFound,
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
