package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/domain"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
)

type CreateContestRepositoryMock struct {
	command.Repository
	isCalled         bool
	result           *command.CreateContestResponse
	err              error
	userContestCount int32
}

func (r *CreateContestRepositoryMock) CreateContest(context.Context, *command.CreateContestRequest) (*command.CreateContestResponse, error) {
	r.isCalled = true
	return r.result, r.err
}

func (r *CreateContestRepositoryMock) GetContestsByUserCountForYear(context.Context, time.Time, uuid.UUID) (int32, error) {
	return r.userContestCount, nil
}

func (r *CreateContestRepositoryMock) UpsertUser(ctx context.Context, req *immersiondomain.UserUpsertRequest) error {
	return nil
}

func (r *CreateContestRepositoryMock) DetachLogFromContest(ctx context.Context, req *immersiondomain.ContestModerationDetachLogRequest, userID uuid.UUID) error {
	return nil
}

func (r *CreateContestRepositoryMock) UpsertContestRegistration(ctx context.Context, req *immersiondomain.RegistrationUpsertRequest) error {
	return nil
}

func (r *CreateContestRepositoryMock) CreateLog(ctx context.Context, req *immersiondomain.LogCreateRequest) (*uuid.UUID, error) {
	return nil, nil
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
		name             string
		request          *command.CreateContestRequest
		role             domain.Role
		err              error
		userContestcount int32
	}{
		{
			"happy path",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			nil,
			0,
		}, {
			"official round cannot be private",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 true,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			command.ErrInvalidContest,
			0,
		}, {
			"official round cannot restrict language choice",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Title:                   "test round",
				LanguageCodeAllowList:   []string{"jpa"},
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			command.ErrInvalidContest,
			0,
		}, {
			"contest cannot be in the past",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(-30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(0),
				RegistrationEnd:         clock.Now().Add(-4 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			command.ErrInvalidContest,
			0,
		}, {
			"admins can bypass contest cannot be in the past",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(-30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(0),
				RegistrationEnd:         clock.Now().Add(-4 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleAdmin,
			nil,
			0,
		}, {
			"needs to start before ending",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(45 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(30 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(10 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			command.ErrInvalidContest,
			0,
		}, {
			"user can create non-official contest",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			nil,
			11,
		}, {
			"user cannot create official contest",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                true,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			command.ErrForbidden,
			0,
		}, {
			"non-official contest can restrict languages",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
				LanguageCodeAllowList:   []string{"jpa"},
			},
			domain.RoleUser,
			nil,
			0,
		}, {
			"user can not create more than 12 contests",
			&command.CreateContestRequest{
				ContestStart:            clock.Now().Add(30 * 24 * time.Hour),
				ContestEnd:              clock.Now().Add(45 * 24 * time.Hour),
				RegistrationEnd:         clock.Now().Add(40 * 24 * time.Hour),
				Official:                false,
				Private:                 false,
				Title:                   "test round",
				ActivityTypeIDAllowList: []int32{1, 2},
			},
			domain.RoleUser,
			command.ErrForbidden,
			12,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := createContext(test.role)
			repo := &CreateContestRepositoryMock{userContestCount: test.userContestcount}
			service := command.NewService(repo, clock)

			_, err := service.CreateContest(ctx, test.request)

			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.err == nil, repo.isCalled)
		})
	}
}
