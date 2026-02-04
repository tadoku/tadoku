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
		role                   commondomain.Role
		requestPageSize        int
		repoResponse           *domain.ContestListResponse
		repoErr                error
		expectedErr            error
		expectedPageSize       int
		expectedIncludePrivate bool
	}{
		{
			name:            "default page size is 10",
			role:            commondomain.RoleUser,
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
			role:            commondomain.RoleUser,
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
			role:            commondomain.RoleUser,
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
			role:            commondomain.RoleAdmin,
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

			repo := &contestListRepositoryMock{
				response: test.repoResponse,
				err:      test.repoErr,
			}
			service := domain.NewContestList(repo)

			result, err := service.Execute(ctx, &domain.ContestListRequest{
				PageSize: test.requestPageSize,
			})

			if test.expectedErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
			assert.Equal(t, test.expectedIncludePrivate, repo.capturedRequest.IncludePrivate)
		})
	}
}
