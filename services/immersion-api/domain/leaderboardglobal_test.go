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

type leaderboardGlobalRepositoryMock struct {
	leaderboard     *domain.Leaderboard
	err             error
	capturedRequest *domain.LeaderboardGlobalRequest

	allScores    []domain.LeaderboardScore
	allScoresErr error
	displayNames map[uuid.UUID]string
	displayErr   error
}

func (m *leaderboardGlobalRepositoryMock) FetchGlobalLeaderboard(ctx context.Context, req *domain.LeaderboardGlobalRequest) (*domain.Leaderboard, error) {
	m.capturedRequest = req
	return m.leaderboard, m.err
}

func (m *leaderboardGlobalRepositoryMock) FindUserDisplayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	return m.displayNames, m.displayErr
}

func (m *leaderboardGlobalRepositoryMock) FetchAllGlobalLeaderboardScores(ctx context.Context) ([]domain.LeaderboardScore, error) {
	return m.allScores, m.allScoresErr
}

type leaderboardGlobalStoreMock struct {
	page     *domain.LeaderboardPage
	exists   bool
	fetchErr error

	rebuildErr    error
	rebuiltScores []domain.LeaderboardScore
}

func (m *leaderboardGlobalStoreMock) FetchGlobalLeaderboardPage(ctx context.Context, page, pageSize int) (*domain.LeaderboardPage, bool, error) {
	return m.page, m.exists, m.fetchErr
}

func (m *leaderboardGlobalStoreMock) RebuildGlobalLeaderboard(ctx context.Context, scores []domain.LeaderboardScore) error {
	m.rebuiltScores = scores
	return m.rebuildErr
}

func TestLeaderboardGlobal_Execute(t *testing.T) {
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
			repo := &leaderboardGlobalRepositoryMock{
				leaderboard: test.repoLeaderboard,
				err:         test.repoErr,
			}
			store := &leaderboardGlobalStoreMock{}
			service := domain.NewLeaderboardGlobal(repo, store)

			// Use a filter so it goes through the repo path
			result, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
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
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
		})
	}
}

func TestLeaderboardGlobal_Filters(t *testing.T) {
	languageCode := "jpn"
	activityID := int32(1)

	validLeaderboard := &domain.Leaderboard{
		Entries:       []domain.LeaderboardEntry{},
		TotalSize:     0,
		NextPageToken: "",
	}

	repo := &leaderboardGlobalRepositoryMock{
		leaderboard: validLeaderboard,
	}
	store := &leaderboardGlobalStoreMock{}
	service := domain.NewLeaderboardGlobal(repo, store)

	_, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
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

func TestLeaderboardGlobal_CacheHit(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()

	store := &leaderboardGlobalStoreMock{
		page: &domain.LeaderboardPage{
			Scores: []domain.LeaderboardScore{
				{UserID: u1, Score: 200},
				{UserID: u2, Score: 100},
			},
			TotalCount: 2,
			StartRank:  1,
		},
		exists: true,
	}
	repo := &leaderboardGlobalRepositoryMock{
		displayNames: map[uuid.UUID]string{u1: "Alice", u2: "Bob"},
	}
	service := domain.NewLeaderboardGlobal(repo, store)

	result, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
		PageSize: 25,
	})

	require.NoError(t, err)
	require.Len(t, result.Entries, 2)
	assert.Equal(t, "Alice", result.Entries[0].UserDisplayName)
	assert.Equal(t, float32(200), result.Entries[0].Score)
	assert.Equal(t, 1, result.Entries[0].Rank)
	assert.Equal(t, "Bob", result.Entries[1].UserDisplayName)
	assert.Equal(t, 2, result.Entries[1].Rank)
	assert.Equal(t, 2, result.TotalSize)
	assert.Nil(t, repo.capturedRequest, "should not call repo.FetchGlobalLeaderboard on cache hit")
}

func TestLeaderboardGlobal_CacheMiss(t *testing.T) {
	validLeaderboard := &domain.Leaderboard{
		Entries:   []domain.LeaderboardEntry{},
		TotalSize: 0,
	}
	allScores := []domain.LeaderboardScore{
		{UserID: uuid.New(), Score: 100},
	}

	store := &leaderboardGlobalStoreMock{
		exists: false,
	}
	repo := &leaderboardGlobalRepositoryMock{
		leaderboard: validLeaderboard,
		allScores:   allScores,
	}
	service := domain.NewLeaderboardGlobal(repo, store)

	result, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
		PageSize: 25,
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, allScores, store.rebuiltScores, "should rebuild store with all scores")
	assert.NotNil(t, repo.capturedRequest, "should fall back to repo.FetchGlobalLeaderboard")
}

func TestLeaderboardGlobal_StoreFetchError(t *testing.T) {
	store := &leaderboardGlobalStoreMock{
		fetchErr: errors.New("valkey down"),
	}
	repo := &leaderboardGlobalRepositoryMock{}
	service := domain.NewLeaderboardGlobal(repo, store)

	_, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
		PageSize: 25,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valkey down")
}

func TestLeaderboardGlobal_Pagination(t *testing.T) {
	u1 := uuid.New()

	store := &leaderboardGlobalStoreMock{
		page: &domain.LeaderboardPage{
			Scores:     []domain.LeaderboardScore{{UserID: u1, Score: 50}},
			TotalCount: 30,
			StartRank:  1,
		},
		exists: true,
	}
	repo := &leaderboardGlobalRepositoryMock{
		displayNames: map[uuid.UUID]string{u1: "A"},
	}
	service := domain.NewLeaderboardGlobal(repo, store)

	result, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{
		PageSize: 10,
		Page:     0,
	})

	require.NoError(t, err)
	assert.Equal(t, "1", result.NextPageToken)
	assert.Equal(t, 30, result.TotalSize)
}
