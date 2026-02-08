package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type registrationListYearlyRepositoryMock struct {
	registrations   *domain.ContestRegistrations
	err             error
	capturedRequest *domain.RegistrationListYearlyRequest
}

func (m *registrationListYearlyRepositoryMock) YearlyContestRegistrationsForUser(ctx context.Context, req *domain.RegistrationListYearlyRequest) (*domain.ContestRegistrations, error) {
	m.capturedRequest = req
	return m.registrations, m.err
}

func TestRegistrationListYearly_Execute(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()

	validRegistrations := &domain.ContestRegistrations{
		Registrations: []domain.ContestRegistration{
			{
				ID:              uuid.New(),
				ContestID:       uuid.New(),
				UserID:          userID,
				UserDisplayName: "TestUser",
				Languages: []domain.Language{
					{Code: "jpn", Name: "Japanese"},
				},
			},
		},
		TotalSize:     1,
		NextPageToken: "",
	}

	tests := []struct {
		name                 string
		admin                bool
		noSession            bool
		sessionUserID        uuid.UUID
		requestUserID        uuid.UUID
		repoRegistrations    *domain.ContestRegistrations
		repoErr              error
		expectedErr          error
		expectRepoCalled     bool
		expectIncludePrivate bool
	}{
		{
			name:                 "user can list own registrations with private",
			sessionUserID:        userID,
			requestUserID:        userID,
			repoRegistrations:    validRegistrations,
			expectRepoCalled:     true,
			expectIncludePrivate: true,
		},
		{
			name:                 "user cannot see private for other users",
			sessionUserID:        userID,
			requestUserID:        otherUserID,
			repoRegistrations:    validRegistrations,
			expectRepoCalled:     true,
			expectIncludePrivate: false,
		},
		{
			name:                 "admin can see private for any user",
			admin:                true,
			sessionUserID:        userID,
			requestUserID:        otherUserID,
			repoRegistrations:    validRegistrations,
			expectRepoCalled:     true,
			expectIncludePrivate: true,
		},
		{
			name:             "guest cannot list registrations",
			noSession:        true,
			sessionUserID:    userID,
			requestUserID:    userID,
			expectedErr:      domain.ErrUnauthorized,
			expectRepoCalled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ctx context.Context
			if test.noSession {
				ctx = context.Background() // No session for guest
			} else if test.admin {
				ctx = ctxWithAdminSubject(test.sessionUserID.String())
			} else {
				ctx = ctxWithUserSubject(test.sessionUserID.String())
			}

			repo := &registrationListYearlyRepositoryMock{
				registrations: test.repoRegistrations,
				err:           test.repoErr,
			}
			service := domain.NewRegistrationListYearly(repo)

			result, err := service.Execute(ctx, &domain.RegistrationListYearlyRequest{
				UserID: test.requestUserID,
				Year:   2024,
			})

			if test.expectedErr != nil {
				assert.ErrorIs(t, err, test.expectedErr)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if test.expectRepoCalled {
				assert.NotNil(t, repo.capturedRequest)
				assert.Equal(t, test.requestUserID, repo.capturedRequest.UserID)
				assert.Equal(t, 2024, repo.capturedRequest.Year)
				assert.Equal(t, test.expectIncludePrivate, repo.capturedRequest.IncludePrivate)
			}
		})
	}
}

func TestRegistrationListYearly_NoSession(t *testing.T) {
	ctx := context.Background() // No session

	repo := &registrationListYearlyRepositoryMock{}
	service := domain.NewRegistrationListYearly(repo)

	result, err := service.Execute(ctx, &domain.RegistrationListYearlyRequest{
		UserID: uuid.New(),
		Year:   2024,
	})

	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Nil(t, result)
}
