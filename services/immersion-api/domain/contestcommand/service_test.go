package contestcommand_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
)

type ContestRepositoryMock struct {
	contestcommand.ContestRepository
	isCalled bool
	result   *contestcommand.ContestCreateResponse
	err      error
}

func (r *ContestRepositoryMock) CreateContest(context.Context, *contestcommand.ContestCreateRequest) (*contestcommand.ContestCreateResponse, error) {
	r.isCalled = true
	return r.result, r.err
}

func createContext(role domain.Role) context.Context {
	token := &domain.SessionToken{
		Role:        role,
		Subject:     uuid.NewString(),
		DisplayName: "Luffy",
	}

	return context.WithValue(context.Background(), domain.CtxSessionKey, token)
}

func TestCreateContest(t *testing.T) {
	clock := domain.NewMockClock(time.Now())

	tests := []struct {
		name    string
		request *contestcommand.ContestCreateRequest
		role    domain.Role
		err     error
	}{
		{
			"happy path",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			nil,
		}, {
			"official round cannot be private",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 true,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			contestcommand.ErrInvalidContest,
		}, {
			"official round cannot restrict language choice",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				LanguageCodeAllowList:   []string{"jpa"},
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			contestcommand.ErrInvalidContest,
		}, {
			"contest cannot be in the past",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(-30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(-45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(-1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(-40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			contestcommand.ErrInvalidContest,
		}, {
			"registration should open before contest starts",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(60 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(80 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			contestcommand.ErrInvalidContest,
		}, {
			"needs to start before ending",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(45 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(30 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(10 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			contestcommand.ErrInvalidContest,
		}, {
			"user can create non-official contest",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			nil,
		}, {
			"user cannot create official contest",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			contestcommand.ErrForbidden,
		}, {
			"non-official contest can restrict languages",
			&contestcommand.ContestCreateRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationStart:       clock.Now().Add(1 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Description:             "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
				LanguageCodeAllowList:   []string{"jpa"},
			},
			domain.RoleUser,
			nil,
		}, {
			"banned user cannot a contest",
			&contestcommand.ContestCreateRequest{},
			domain.RoleBanned,
			contestcommand.ErrForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := createContext(test.role)
			repo := &ContestRepositoryMock{}
			service := contestcommand.NewService(repo, clock)

			_, err := service.CreateContest(ctx, test.request)

			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.err == nil, repo.isCalled)
		})
	}
}
