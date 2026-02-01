package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type profileYearlyScoresRepositoryMock struct {
	scores          []domain.Score
	err             error
	capturedRequest *domain.ProfileYearlyScoresRequest
}

func (m *profileYearlyScoresRepositoryMock) YearlyScoresForUser(ctx context.Context, req *domain.ProfileYearlyScoresRequest) ([]domain.Score, error) {
	m.capturedRequest = req
	return m.scores, m.err
}

func TestProfileYearlyScores_Execute(t *testing.T) {
	userID := uuid.New()
	jpnName := "Japanese"
	engName := "English"

	validScores := []domain.Score{
		{LanguageCode: "jpn", LanguageName: &jpnName, Score: 100.5},
		{LanguageCode: "eng", LanguageName: &engName, Score: 50.0},
	}

	tests := []struct {
		name                 string
		scores               []domain.Score
		repoErr              error
		expectedOverallScore float32
		expectedErr          bool
	}{
		{
			name:                 "successful scores fetch",
			scores:               validScores,
			expectedOverallScore: 150.5,
		},
		{
			name:        "repository error",
			repoErr:     errors.New("database error"),
			expectedErr: true,
		},
		{
			name:                 "empty scores",
			scores:               []domain.Score{},
			expectedOverallScore: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &profileYearlyScoresRepositoryMock{
				scores: test.scores,
				err:    test.repoErr,
			}
			service := domain.NewProfileYearlyScores(repo)

			result, err := service.Execute(context.Background(), &domain.ProfileYearlyScoresRequest{
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
			assert.Equal(t, test.expectedOverallScore, result.OverallScore)

			// Verify request was passed correctly
			assert.NotNil(t, repo.capturedRequest)
			assert.Equal(t, userID, repo.capturedRequest.UserID)
			assert.Equal(t, 2024, repo.capturedRequest.Year)
		})
	}
}
