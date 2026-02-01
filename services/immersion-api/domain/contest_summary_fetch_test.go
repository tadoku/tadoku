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

type mockContestSummaryFetchRepo struct {
	fetchFn func(ctx context.Context, req *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error)
}

func (m *mockContestSummaryFetchRepo) FetchContestSummary(ctx context.Context, req *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx, req)
	}
	return nil, nil
}

func TestContestSummaryFetch_Execute(t *testing.T) {
	t.Run("fetches contest summary successfully", func(t *testing.T) {
		contestID := uuid.New()

		repo := &mockContestSummaryFetchRepo{
			fetchFn: func(ctx context.Context, req *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error) {
				if req.ContestID == contestID {
					return &domain.ContestSummaryFetchResponse{
						ParticipantCount: 42,
						LanguageCount:    5,
						TotalScore:       12345.67,
					}, nil
				}
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewContestSummaryFetch(repo)
		resp, err := svc.Execute(context.Background(), &domain.ContestSummaryFetchRequest{
			ContestID: contestID,
		})

		require.NoError(t, err)
		assert.Equal(t, 42, resp.ParticipantCount)
		assert.Equal(t, 5, resp.LanguageCount)
		assert.Equal(t, float32(12345.67), resp.TotalScore)
	})

	t.Run("returns not found when contest does not exist", func(t *testing.T) {
		repo := &mockContestSummaryFetchRepo{
			fetchFn: func(ctx context.Context, req *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error) {
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewContestSummaryFetch(repo)
		_, err := svc.Execute(context.Background(), &domain.ContestSummaryFetchRequest{
			ContestID: uuid.New(),
		})

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockContestSummaryFetchRepo{
			fetchFn: func(ctx context.Context, req *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewContestSummaryFetch(repo)
		_, err := svc.Execute(context.Background(), &domain.ContestSummaryFetchRequest{
			ContestID: uuid.New(),
		})

		assert.ErrorIs(t, err, repoErr)
	})
}
