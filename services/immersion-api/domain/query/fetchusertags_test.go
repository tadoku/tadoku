package query_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type FetchUserTagsRepositoryMock struct {
	query.Repository
	tags *query.FetchUserTagsResponse
	err  error
}

func (r *FetchUserTagsRepositoryMock) FetchUserTags(ctx context.Context, req *query.FetchUserTagsRequest) (*query.FetchUserTagsResponse, error) {
	return r.tags, r.err
}

func createContextWithSession(userID uuid.UUID) context.Context {
	token := &domain.SessionToken{
		Role:        domain.RoleUser,
		Subject:     userID.String(),
		DisplayName: "TestUser",
	}
	return context.WithValue(context.Background(), domain.CtxSessionKey, token)
}

func TestFetchUserTags(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()

	mockTags := &query.FetchUserTagsResponse{
		Tags: []query.UserTag{
			{Tag: "reading", UsageCount: 5},
			{Tag: "vn", UsageCount: 3},
		},
		DefaultTags: []string{"reading"},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		requestUser uuid.UUID
		expectError error
		expectTags  *query.FetchUserTagsResponse
	}{
		{
			name:        "authorized - same user",
			ctx:         createContextWithSession(userA),
			requestUser: userA,
			expectError: nil,
			expectTags:  mockTags,
		},
		{
			name:        "unauthorized - no session",
			ctx:         context.Background(),
			requestUser: userA,
			expectError: query.ErrUnauthorized,
			expectTags:  nil,
		},
		{
			name:        "unauthorized - different user",
			ctx:         createContextWithSession(userB),
			requestUser: userA,
			expectError: query.ErrUnauthorized,
			expectTags:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &FetchUserTagsRepositoryMock{tags: mockTags}
			service := query.NewService(repo, nil, nil)

			result, err := service.FetchUserTags(test.ctx, &query.FetchUserTagsRequest{
				UserID: test.requestUser,
			})

			if test.expectError != nil {
				assert.ErrorIs(t, err, test.expectError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectTags, result)
			}
		})
	}
}
