package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type registrationFindRepositoryMock struct {
	registration    *domain.ContestRegistration
	err             error
	capturedRequest *domain.RegistrationFindRequest
}

func (m *registrationFindRepositoryMock) FindRegistrationForUser(ctx context.Context, req *domain.RegistrationFindRequest) (*domain.ContestRegistration, error) {
	m.capturedRequest = req
	return m.registration, m.err
}

func TestRegistrationFind_Execute(t *testing.T) {
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

	tests := []struct {
		name             string
		admin            bool
		guest            bool
		userID           uuid.UUID
		repoRegistration *domain.ContestRegistration
		repoErr          error
		expectedErr      error
		expectRepoCalled bool
	}{
		{
			name:             "authenticated user can find registration",
			userID:           userID,
			repoRegistration: validRegistration,
			expectRepoCalled: true,
		},
		{
			name:             "admin can find registration",
			admin:            true,
			userID:           userID,
			repoRegistration: validRegistration,
			expectRepoCalled: true,
		},
		{
			name:             "guest cannot find registration",
			guest:            true,
			userID:           userID,
			expectedErr:      domain.ErrUnauthorized,
			expectRepoCalled: false,
		},
		{
			name:             "registration not found",
			userID:           userID,
			repoErr:          domain.ErrNotFound,
			expectedErr:      domain.ErrNotFound,
			expectRepoCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ctx context.Context
			if test.guest {
				ctx = ctxWithGuest()
			} else if test.admin {
				ctx = ctxWithAdminSubject(test.userID.String())
			} else {
				ctx = ctxWithUserSubject(test.userID.String())
			}

			repo := &registrationFindRepositoryMock{
				registration: test.repoRegistration,
				err:          test.repoErr,
			}
			service := domain.NewRegistrationFind(repo)

			result, err := service.Execute(ctx, &domain.RegistrationFindRequest{
				ContestID: contestID,
			})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if test.expectRepoCalled {
				assert.NotNil(t, repo.capturedRequest)
				assert.Equal(t, test.userID, repo.capturedRequest.UserID)
				assert.Equal(t, contestID, repo.capturedRequest.ContestID)
			}
		})
	}
}

func TestRegistrationFind_NoSession(t *testing.T) {
	ctx := context.Background() // No session

	repo := &registrationFindRepositoryMock{}
	service := domain.NewRegistrationFind(repo)

	result, err := service.Execute(ctx, &domain.RegistrationFindRequest{
		ContestID: uuid.New(),
	})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Nil(t, result)
}
