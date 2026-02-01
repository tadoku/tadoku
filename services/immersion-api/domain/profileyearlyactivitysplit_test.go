package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockProfileYearlyActivitySplitRepo struct {
	fetchFn func(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error)
}

func (m *mockProfileYearlyActivitySplitRepo) YearlyActivitySplitForUser(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx, req)
	}
	return nil, nil
}

func TestProfileYearlyActivitySplit_Execute(t *testing.T) {
	t.Run("fetches yearly activity split successfully", func(t *testing.T) {
		userID := uuid.New()

		repo := &mockProfileYearlyActivitySplitRepo{
			fetchFn: func(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error) {
				if req.UserID == userID && req.Year == 2024 {
					return &domain.ProfileYearlyActivitySplitResponse{
						Activities: []domain.ActivityScore{
							{ActivityID: 1, ActivityName: "Reading", Score: 1000.5},
							{ActivityID: 2, ActivityName: "Listening", Score: 500.25},
						},
					}, nil
				}
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewProfileYearlyActivitySplit(repo)
		resp, err := svc.Execute(context.Background(), &domain.ProfileYearlyActivitySplitRequest{
			UserID: userID,
			Year:   2024,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Activities, 2)
		assert.Equal(t, "Reading", resp.Activities[0].ActivityName)
		assert.Equal(t, float32(1000.5), resp.Activities[0].Score)
	})

	t.Run("returns empty activities when user has no logs", func(t *testing.T) {
		repo := &mockProfileYearlyActivitySplitRepo{
			fetchFn: func(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error) {
				return &domain.ProfileYearlyActivitySplitResponse{
					Activities: []domain.ActivityScore{},
				}, nil
			},
		}

		svc := domain.NewProfileYearlyActivitySplit(repo)
		resp, err := svc.Execute(context.Background(), &domain.ProfileYearlyActivitySplitRequest{
			UserID: uuid.New(),
			Year:   2024,
		})

		require.NoError(t, err)
		assert.Empty(t, resp.Activities)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockProfileYearlyActivitySplitRepo{
			fetchFn: func(ctx context.Context, req *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewProfileYearlyActivitySplit(repo)
		_, err := svc.Execute(context.Background(), &domain.ProfileYearlyActivitySplitRequest{
			UserID: uuid.New(),
			Year:   2024,
		})

		assert.ErrorIs(t, err, repoErr)
	})
}
