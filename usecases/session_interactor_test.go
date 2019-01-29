package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setup(t *testing.T) (
	*gomock.Controller,
	*usecases.MockUserRepository,
	*usecases.MockPasswordHasher,
	*usecases.MockJWTGenerator,
	usecases.SessionInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockUserRepository(ctrl)
	pwHasher := usecases.NewMockPasswordHasher(ctrl)
	jwtGen := usecases.NewMockJWTGenerator(ctrl)

	interactor := usecases.NewSessionInteractor(
		repo, pwHasher, jwtGen, time.Hour*1,
	)

	return ctrl, repo, pwHasher, jwtGen, interactor
}

func TestSessionInteractor_CreateUser(t *testing.T) {
	ctrl, repo, pwHasher, _, interactor := setup(t)
	defer ctrl.Finish()

	user := domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}
	hashedUser := user
	hashedUser.Password = "barbar"

	pwHasher.EXPECT().Hash(user.Password).Return(hashedUser.Password, nil)
	repo.EXPECT().Store(hashedUser)

	err := interactor.CreateUser(user)

	assert.NoError(t, err)
}

func TestSessionInteractor_CreateSession(t *testing.T) {
	ctrl, repo, pwHasher, jwtGen, interactor := setup(t)
	defer ctrl.Finish()

	{
		// Happy path: valid user
		dbUser := domain.User{ID: 1, Email: "foo@bar.com", Password: "foobar"}
		repo.EXPECT().FindByEmail("foo@bar.com").Return(dbUser, nil)
		pwHasher.EXPECT().Compare(dbUser.Password, "foobar").Return(true)
		jwtGen.EXPECT().NewToken(gomock.Any(), usecases.SessionClaims{User: &dbUser}).Return("token", nil)

		sessionUser, token, err := interactor.CreateSession("foo@bar.com", "foobar")
		assert.NoError(t, err)
		assert.Equal(t, sessionUser, dbUser)
		assert.Equal(t, token, "token")
	}

	{
		// Sad path: user does not exist
		repo.EXPECT().FindByEmail("bar@bar.com").Return(domain.User{}, nil)
		_, _, err := interactor.CreateSession("bar@bar.com", "foobar")
		assert.EqualError(t, err, usecases.ErrUserDoesNotExist.Error())
	}

	{
		// Sad path: password is incorrect
		user := domain.User{ID: 1, Email: "foo@bar.com", Password: "barbar"}
		repo.EXPECT().FindByEmail("foo@bar.com").Return(user, nil)
		pwHasher.EXPECT().Compare(user.Password, "foobar").Return(false)
		_, _, err := interactor.CreateSession("foo@bar.com", "foobar")
		assert.EqualError(t, err, usecases.ErrPasswordIncorrect.Error())
	}
}
