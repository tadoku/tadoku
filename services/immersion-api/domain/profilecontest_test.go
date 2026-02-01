package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type profileContestRepositoryMock struct {
	registration      *domain.ContestRegistration
	registrationErr   error
	scores            []domain.Score
	scoresErr         error
	capturedRegReq    *domain.RegistrationFindRequest
	capturedScoresReq *domain.ProfileContestRequest
}

func (m *profileContestRepositoryMock) FindRegistrationForUser(ctx context.Context, req *domain.RegistrationFindRequest) (*domain.ContestRegistration, error) {
	m.capturedRegReq = req
	return m.registration, m.registrationErr
}

func (m *profileContestRepositoryMock) FindScoresForRegistration(ctx context.Context, req *domain.ProfileContestRequest) ([]domain.Score, error) {
	m.capturedScoresReq = req
	return m.scores, m.scoresErr
}

func TestProfileContest_Execute(t *testing.T) {
	userID := uuid.New()
	contestID := uuid.New()

	validRegistration := &domain.ContestRegistration{
		ID:              uuid.New(),
		ContestID:       contestID,
		UserID:          userID,
		UserDisplayName: "TestUser",
		Languages: []domain.Language{
			{Code: "jpn", Name: "Japanese"},
		},
	}

	validScores := []domain.Score{
		{LanguageCode: "jpn", Score: 100.5},
		{LanguageCode: "eng", Score: 50.0},
	}

	tests := []struct {
		name                 string
		registration         *domain.ContestRegistration
		registrationErr      error
		scores               []domain.Score
		scoresErr            error
		expectedOverallScore float32
		expectedErr          bool
		expectedErrSubstring string
	}{
		{
			name:                 "successful profile fetch",
			registration:         validRegistration,
			scores:               validScores,
			expectedOverallScore: 150.5,
		},
		{
			name:                 "registration not found",
			registrationErr:      domain.ErrNotFound,
			expectedErr:          true,
			expectedErrSubstring: "could not fetch registration",
		},
		{
			name:                 "scores error",
			registration:         validRegistration,
			scoresErr:            errors.New("database error"),
			expectedErr:          true,
			expectedErrSubstring: "could not fetch scores",
		},
		{
			name:                 "empty scores",
			registration:         validRegistration,
			scores:               []domain.Score{},
			expectedOverallScore: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &profileContestRepositoryMock{
				registration:    test.registration,
				registrationErr: test.registrationErr,
				scores:          test.scores,
				scoresErr:       test.scoresErr,
			}
			service := domain.NewProfileContest(repo)

			result, err := service.Execute(context.Background(), &domain.ProfileContestRequest{
				UserID:    userID,
				ContestID: contestID,
			})

			if test.expectedErr {
				assert.Error(t, err)
				if test.expectedErrSubstring != "" {
					assert.Contains(t, err.Error(), test.expectedErrSubstring)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, test.expectedOverallScore, result.OverallScore)
			assert.Equal(t, test.registration, result.Registration)
			assert.Equal(t, len(test.scores), len(result.Scores))

			// Verify registration request was passed correctly
			assert.NotNil(t, repo.capturedRegReq)
			assert.Equal(t, userID, repo.capturedRegReq.UserID)
			assert.Equal(t, contestID, repo.capturedRegReq.ContestID)
		})
	}
}
