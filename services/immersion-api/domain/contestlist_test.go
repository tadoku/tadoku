package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

var errContestListDatabase = errors.New("database error")

type contestListRepositoryMock struct {
	response        *domain.ContestListResponse
	err             error
	capturedRequest *domain.ContestListRequest
}

func (m *contestListRepositoryMock) ListContests(ctx context.Context, req *domain.ContestListRequest) (*domain.ContestListResponse, error) {
	m.capturedRequest = req
	return m.response, m.err
}

func TestContestList_Execute(t *testing.T) {
	tests := []struct {
		name                   string
		admin                  bool
		requestPageSize        int
		repoResponse           *domain.ContestListResponse
		repoErr                error
		expectedErr            error
		expectedPageSize       int
		expectedIncludePrivate bool
	}{
		{
			name:            "default page size is 10",
			admin:           false,
			requestPageSize: 0,
			repoResponse: &domain.ContestListResponse{
				Contests:  []domain.Contest{},
				TotalSize: 0,
			},
			expectedPageSize:       10,
			expectedIncludePrivate: false,
		},
		{
			name:            "page size capped at 100",
			admin:           false,
			requestPageSize: 500,
			repoResponse: &domain.ContestListResponse{
				Contests:  []domain.Contest{},
				TotalSize: 0,
			},
			expectedPageSize:       100,
			expectedIncludePrivate: false,
		},
		{
			name:            "custom page size is preserved",
			admin:           false,
			requestPageSize: 25,
			repoResponse: &domain.ContestListResponse{
				Contests:  []domain.Contest{},
				TotalSize: 0,
			},
			expectedPageSize:       25,
			expectedIncludePrivate: false,
		},
		{
			name:            "admin can see private contests",
			admin:           true,
			requestPageSize: 10,
			repoResponse: &domain.ContestListResponse{
				Contests:  []domain.Contest{},
				TotalSize: 0,
			},
			expectedPageSize:       10,
			expectedIncludePrivate: true,
		},
		{
			name:            "repository error is propagated",
			admin:           false,
			requestPageSize: 10,
			repoErr:         errContestListDatabase,
			expectedErr:     errContestListDatabase,
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

			repo := &contestListRepositoryMock{
				response: test.repoResponse,
				err:      test.repoErr,
			}
			service := domain.NewContestList(repo)

			result, err := service.Execute(ctx, &domain.ContestListRequest{
				PageSize: test.requestPageSize,
			})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
			assert.Equal(t, test.expectedIncludePrivate, repo.capturedRequest.IncludePrivate)
		})
	}
}
