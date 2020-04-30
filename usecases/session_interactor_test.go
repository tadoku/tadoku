package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

var sessionLength = time.Hour * 1

func setupSessionTest(t *testing.T) (
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
		repo, pwHasher, jwtGen, sessionLength,
	)

	return ctrl, repo, pwHasher, jwtGen, interactor
}

func TestSessionInteractor_CreateSession(t *testing.T) {
	ctrl, repo, pwHasher, jwtGen, interactor := setupSessionTest(t)
	defer ctrl.Finish()

	{
		// Happy path: valid user
		dbUser := domain.User{ID: 1, Email: "foo@bar.com", Password: "foobar"}
		repo.EXPECT().FindByEmail("foo@bar.com").Return(dbUser, nil)
		pwHasher.EXPECT().Compare(dbUser.Password, "foobar").Return(true)
		expectedExpiration := time.Now().Unix()
		jwtGen.EXPECT().NewToken(sessionLength, usecases.SessionClaims{User: &dbUser}).Return("token", expectedExpiration, nil)

		sessionUser, token, expiresAt, err := interactor.CreateSession("foo@bar.com", "foobar")
		assert.NoError(t, err)
		assert.Equal(t, sessionUser, dbUser)
		assert.Equal(t, token, "token")
		assert.Equal(t, expiresAt, expectedExpiration)
	}

	{
		// Sad path: user does not exist
		repo.EXPECT().FindByEmail("bar@bar.com").Return(domain.User{}, nil)
		_, _, _, err := interactor.CreateSession("bar@bar.com", "foobar")
		assert.EqualError(t, err, usecases.ErrUserDoesNotExist.Error())
	}

	{
		// Sad path: password is incorrect
		user := domain.User{ID: 1, Email: "foo@bar.com", Password: "barbar"}
		repo.EXPECT().FindByEmail("foo@bar.com").Return(user, nil)
		pwHasher.EXPECT().Compare(user.Password, "foobar").Return(false)
		_, _, _, err := interactor.CreateSession("foo@bar.com", "foobar")
		assert.EqualError(t, err, domain.ErrPasswordIncorrect.Error())
	}
}

func TestSessionInteractor_RefreshSession(t *testing.T) {
	ctrl, repo, _, jwtGen, interactor := setupSessionTest(t)
	defer ctrl.Finish()

	{
		user := domain.User{ID: 1, DisplayName: "foo", Email: "foo@bar.com", Password: "foobar"}
		dbUser := domain.User{ID: 1, DisplayName: "bar", Email: "foo@bar.com", Password: "foobar"}

		repo.EXPECT().FindByEmail("foo@bar.com").Return(dbUser, nil)
		expectedExpiration := time.Now().Unix()
		jwtGen.EXPECT().NewToken(sessionLength, usecases.SessionClaims{User: &dbUser}).Return("token", expectedExpiration, nil)

		sessionUser, token, expiresAt, err := interactor.RefreshSession(user)
		assert.NoError(t, err)
		assert.Equal(t, sessionUser, dbUser)
		assert.Equal(t, token, "token")
		assert.Equal(t, expiresAt, expectedExpiration)
	}
}
