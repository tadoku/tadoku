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

type profileContestActivityRepositoryMock struct {
	rows            []domain.ProfileContestActivityRow
	err             error
	capturedRequest *domain.ProfileContestActivityRequest
}

func (m *profileContestActivityRepositoryMock) ActivityForContestUser(ctx context.Context, req *domain.ProfileContestActivityRequest) ([]domain.ProfileContestActivityRow, error) {
	m.capturedRequest = req
	return m.rows, m.err
}

func TestProfileContestActivity_Execute(t *testing.T) {
	userID := uuid.New()
	contestID := uuid.New()

	validRows := []domain.ProfileContestActivityRow{
		{
			Date:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			LanguageCode: "jpn",
			Score:        100.5,
		},
		{
			Date:         time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			LanguageCode: "jpn",
			Score:        50.0,
		},
	}

	tests := []struct {
		name        string
		rows        []domain.ProfileContestActivityRow
		repoErr     error
		expectedErr bool
	}{
		{
			name: "successful activity fetch",
			rows: validRows,
		},
		{
			name:        "repository error",
			repoErr:     errors.New("database error"),
			expectedErr: true,
		},
		{
			name: "empty rows",
			rows: []domain.ProfileContestActivityRow{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &profileContestActivityRepositoryMock{
				rows: test.rows,
				err:  test.repoErr,
			}
			service := domain.NewProfileContestActivity(repo)

			result, err := service.Execute(context.Background(), &domain.ProfileContestActivityRequest{
				UserID:    userID,
				ContestID: contestID,
			})

			if test.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, len(test.rows), len(result.Rows))

			// Verify request was passed correctly
			assert.NotNil(t, repo.capturedRequest)
			assert.Equal(t, userID, repo.capturedRequest.UserID)
			assert.Equal(t, contestID, repo.capturedRequest.ContestID)
		})
	}
}
