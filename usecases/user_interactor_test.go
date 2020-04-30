package usecases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setupUserTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockUserRepository,
	*usecases.MockPasswordHasher,
	usecases.UserInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockUserRepository(ctrl)
	pwHasher := usecases.NewMockPasswordHasher(ctrl)

	interactor := usecases.NewUserInteractor(repo, pwHasher)

	return ctrl, repo, pwHasher, interactor
}

func TestUserInteractor_CreateUser(t *testing.T) {
	ctrl, repo, pwHasher, interactor := setupUserTest(t)
	defer ctrl.Finish()

	user := domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}
	hashedUser := user
	hashedUser.Password = "barbar"

	pwHasher.EXPECT().Hash(user.Password).Return(hashedUser.Password, nil)
	repo.EXPECT().Store(&hashedUser)

	err := interactor.CreateUser(user)

	assert.NoError(t, err)
}

func TestUserInteractor_UpdatePassword(t *testing.T) {
	ctrl, repo, pwHasher, interactor := setupUserTest(t)
	defer ctrl.Finish()

	{
		// Happy path: valid user/password combination
		dbUser := domain.User{ID: 1, Email: "foo@bar.com", Password: "foobar"}
		hashedUser := dbUser
		hashedUser.Password = "foofoo"

		repo.EXPECT().FindByEmail("foo@bar.com").Return(dbUser, nil)
		repo.EXPECT().UpdatePassword(&hashedUser)
		pwHasher.EXPECT().Compare(dbUser.Password, "foobar").Return(true)
		pwHasher.EXPECT().Hash(domain.Password("barbar")).Return(hashedUser.Password, nil)

		err := interactor.UpdatePassword("foo@bar.com", "foobar", "barbar")
		assert.NoError(t, err)
	}

	{
		// Sad path: user does not exist
		repo.EXPECT().FindByEmail("bar@bar.com").Return(domain.User{}, nil)
		err := interactor.UpdatePassword("bar@bar.com", "foobar", "barbar")
		assert.EqualError(t, err, usecases.ErrUserDoesNotExist.Error())
	}

	{
		// Sad path: password is incorrect
		user := domain.User{ID: 1, Email: "foo@bar.com", Password: "barbar"}
		repo.EXPECT().FindByEmail("foo@bar.com").Return(user, nil)
		pwHasher.EXPECT().Compare(user.Password, "foobar").Return(false)
		err := interactor.UpdatePassword("foo@bar.com", "foobar", "foofoo")
		assert.EqualError(t, err, domain.ErrPasswordIncorrect.Error())
	}
}

func TestUserInteractor_UpdateProfile(t *testing.T) {
	ctrl, repo, _, interactor := setupUserTest(t)
	defer ctrl.Finish()

	{
		user := domain.User{ID: 1, DisplayName: "test", Email: "foo@bar.com", Password: ""}

		repo.EXPECT().Store(&user)

		err := interactor.UpdateProfile(user)
		assert.NoError(t, err)
	}
}
