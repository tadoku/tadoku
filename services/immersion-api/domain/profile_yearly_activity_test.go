package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type profileYearlyActivityRepositoryMock struct {
	scores          []domain.UserActivityScore
	err             error
	capturedRequest *domain.ProfileYearlyActivityRequest
}

func (m *profileYearlyActivityRepositoryMock) YearlyActivityForUser(ctx context.Context, req *domain.ProfileYearlyActivityRequest) ([]domain.UserActivityScore, error) {
	m.capturedRequest = req
	return m.scores, m.err
}

func TestProfileYearlyActivity_Execute(t *testing.T) {
	userID := uuid.New()

	validScores := []domain.UserActivityScore{
		{
			Date:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Score:   100.5,
			Updates: 5,
		},
		{
			Date:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Score:   50.0,
			Updates: 3,
		},
	}

	tests := []struct {
		name                 string
		scores               []domain.UserActivityScore
		repoErr              error
		expectedTotalUpdates int
		expectedErr          bool
	}{
		{
			name:                 "successful activity fetch",
			scores:               validScores,
			expectedTotalUpdates: 8,
		},
		{
			name:        "repository error",
			repoErr:     errors.New("database error"),
			expectedErr: true,
		},
		{
			name:                 "empty scores",
			scores:               []domain.UserActivityScore{},
			expectedTotalUpdates: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &profileYearlyActivityRepositoryMock{
				scores: test.scores,
				err:    test.repoErr,
			}
			service := domain.NewProfileYearlyActivity(repo)

			result, err := service.Execute(context.Background(), &domain.ProfileYearlyActivityRequest{
				UserID: userID,
				Year:   2024,
			})

			if test.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, len(test.scores), len(result.Scores))
			assert.Equal(t, test.expectedTotalUpdates, result.TotalUpdates)

			// Verify request was passed correctly
			assert.NotNil(t, repo.capturedRequest)
			assert.Equal(t, userID, repo.capturedRequest.UserID)
			assert.Equal(t, 2024, repo.capturedRequest.Year)
		})
	}
}
