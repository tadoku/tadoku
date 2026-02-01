package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type contestLeaderboardFetchRepositoryMock struct {
	leaderboard     *domain.Leaderboard
	err             error
	capturedRequest *domain.ContestLeaderboardFetchRequest
}

func (m *contestLeaderboardFetchRepositoryMock) FetchContestLeaderboard(ctx context.Context, req *domain.ContestLeaderboardFetchRequest) (*domain.Leaderboard, error) {
	m.capturedRequest = req
	return m.leaderboard, m.err
}

func TestContestLeaderboardFetch_Execute(t *testing.T) {
	contestID := uuid.New()

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &contestLeaderboardFetchRepositoryMock{
				leaderboard: test.repoLeaderboard,
				err:         test.repoErr,
			}
			service := domain.NewContestLeaderboardFetch(repo)

			result, err := service.Execute(context.Background(), &domain.ContestLeaderboardFetchRequest{
				ContestID: contestID,
				PageSize:  test.pageSize,
			})

			if test.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, repo.capturedRequest)
			assert.Equal(t, contestID, repo.capturedRequest.ContestID)
			assert.Equal(t, test.expectedPageSize, repo.capturedRequest.PageSize)
		})
	}
}

func TestContestLeaderboardFetch_Filters(t *testing.T) {
	contestID := uuid.New()
	languageCode := "jpn"
	activityID := int32(1)

	validLeaderboard := &domain.Leaderboard{
		Entries:       []domain.LeaderboardEntry{},
		TotalSize:     0,
		NextPageToken: "",
	}

	repo := &contestLeaderboardFetchRepositoryMock{
		leaderboard: validLeaderboard,
	}
	service := domain.NewContestLeaderboardFetch(repo)

	_, err := service.Execute(context.Background(), &domain.ContestLeaderboardFetchRequest{
		ContestID:    contestID,
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
