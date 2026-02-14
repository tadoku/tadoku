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

type leaderboardYearlyRepositoryMock struct {
	leaderboard     *domain.Leaderboard
	err             error
	capturedRequest *domain.LeaderboardYearlyRequest

	allScores    []domain.LeaderboardScore
	allScoresErr error
	displayNames map[uuid.UUID]string
	displayErr   error
}

func (m *leaderboardYearlyRepositoryMock) FetchYearlyLeaderboard(ctx context.Context, req *domain.LeaderboardYearlyRequest) (*domain.Leaderboard, error) {
	m.capturedRequest = req
	return m.leaderboard, m.err
}

func (m *leaderboardYearlyRepositoryMock) FindUserDisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	return m.displayNames, m.displayErr
}

func (m *leaderboardYearlyRepositoryMock) FetchAllYearlyLeaderboardScores(ctx context.Context, year int) ([]domain.LeaderboardScore, error) {
	return m.allScores, m.allScoresErr
}

type leaderboardYearlyStoreMock struct {
	scores     []domain.LeaderboardScore
	totalCount int
	exists     bool
	fetchErr   error
	rebuildErr error

	rebuiltScores []domain.LeaderboardScore
}

func (m *leaderboardYearlyStoreMock) FetchYearlyLeaderboardPage(ctx context.Context, year int, page, pageSize int) ([]domain.LeaderboardScore, int, bool, error) {
	return m.scores, m.totalCount, m.exists, m.fetchErr
}

func (m *leaderboardYearlyStoreMock) RebuildYearlyLeaderboard(ctx context.Context, year int, scores []domain.LeaderboardScore) error {
	m.rebuiltScores = scores
	return m.rebuildErr
}

func TestLeaderboardYearly_Execute(t *testing.T) {
	validLeaderboard := &domain.Leaderboard{
		Entries: []domain.LeaderboardEntry{
			{
				Rank:            1,
				UserID:          uuid.New(),
				UserDisplayName: "User1",
				Score:           100.5,
				IsTie:           false,
			},
		},
		TotalSize:     1,
		NextPageToken: "",
	}

	tests := []struct {
		name             string
		pageSize         int
		expectedPageSize int
		repoLeaderboard  *domain.Leaderboard
		repoErr          error
		expectedErr      error
	}{
		{
			name:             "default page size when zero",
			pageSize:         0,
			expectedPageSize: 25,
			repoLeaderboard:  validLeaderboard,
		},
		{
			name:             "respects custom page size",
			pageSize:         50,
			expectedPageSize: 50,
			repoLeaderboard:  validLeaderboard,
		},
		{
			name:             "caps page size at 100",
			pageSize:         150,
			expectedPageSize: 100,
			repoLeaderboard:  validLeaderboard,
		},
		{
			name:             "propagates repository error",
			pageSize:         25,
			expectedPageSize: 25,
			repoErr:          domain.ErrNotFound,
			expectedErr:      domain.ErrNotFound,
		},
	}

	languageCode := "jpn"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &leaderboardYearlyRepositoryMock{
				leaderboard: test.repoLeaderboard,
				err:         test.repoErr,
			}
			store := &leaderboardYearlyStoreMock{}
			service := domain.NewLeaderboardYearly(repo, store)

			result, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{
				Year:         2024,
				LanguageCode: &languageCode,
				PageSize:     test.pageSize,
			})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, repo.capturedRequest)
			assert.Equal(t, int32(2024), repo.capturedRequest.Year)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
		})
	}
}

func TestLeaderboardYearly_Filters(t *testing.T) {
	languageCode := "jpn"
	activityID := int32(1)

	validLeaderboard := &domain.Leaderboard{
		Entries:       []domain.LeaderboardEntry{},
		TotalSize:     0,
		NextPageToken: "",
	}

	repo := &leaderboardYearlyRepositoryMock{
		leaderboard: validLeaderboard,
	}
	store := &leaderboardYearlyStoreMock{}
	service := domain.NewLeaderboardYearly(repo, store)

	_, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{
		Year:         2024,
		LanguageCode: &languageCode,
		ActivityID:   &activityID,
		PageSize:     25,
		Page:         2,
	})

	assert.NoError(t, err)
	assert.NotNil(t, repo.capturedRequest)
	assert.Equal(t, &languageCode, repo.capturedRequest.LanguageCode)
	assert.Equal(t, &activityID, repo.capturedRequest.ActivityID)
	assert.Equal(t, 2, repo.capturedRequest.Page)
}

func TestLeaderboardYearly_CacheHit(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()

	store := &leaderboardYearlyStoreMock{
		scores: []domain.LeaderboardScore{
			{UserID: u1, Score: 200},
			{UserID: u2, Score: 100},
		},
		totalCount: 2,
		exists:     true,
	}
	repo := &leaderboardYearlyRepositoryMock{
		displayNames: map[uuid.UUID]string{u1: "Alice", u2: "Bob"},
	}
	service := domain.NewLeaderboardYearly(repo, store)

	result, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{
		Year:     2024,
		PageSize: 25,
	})

	require.NoError(t, err)
	require.Len(t, result.Entries, 2)
	assert.Equal(t, "Alice", result.Entries[0].UserDisplayName)
	assert.Equal(t, float32(200), result.Entries[0].Score)
	assert.Equal(t, 1, result.Entries[0].Rank)
	assert.Nil(t, repo.capturedRequest, "should not call repo.FetchYearlyLeaderboard on cache hit")
}

func TestLeaderboardYearly_CacheMiss(t *testing.T) {
	validLeaderboard := &domain.Leaderboard{
		Entries:   []domain.LeaderboardEntry{},
		TotalSize: 0,
	}
	allScores := []domain.LeaderboardScore{
		{UserID: uuid.New(), Score: 100},
	}

	store := &leaderboardYearlyStoreMock{
		exists: false,
	}
	repo := &leaderboardYearlyRepositoryMock{
		leaderboard: validLeaderboard,
		allScores:   allScores,
	}
	service := domain.NewLeaderboardYearly(repo, store)

	result, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{
		Year:     2024,
		PageSize: 25,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, allScores, store.rebuiltScores, "should rebuild store with all scores")
	assert.NotNil(t, repo.capturedRequest, "should fall back to repo.FetchYearlyLeaderboard")
}

func TestLeaderboardYearly_StoreFetchError(t *testing.T) {
	store := &leaderboardYearlyStoreMock{
		fetchErr: errors.New("valkey down"),
	}
	repo := &leaderboardYearlyRepositoryMock{}
	service := domain.NewLeaderboardYearly(repo, store)

	_, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{
		Year:     2024,
		PageSize: 25,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valkey down")
}
