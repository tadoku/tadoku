package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func TestSessionInteractor_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := usecases.NewMockUserRepository(ctrl)
	pwHasher := usecases.NewMockPasswordHasher(ctrl)
	jwtGen := usecases.NewMockJWTGenerator(ctrl)

	user := domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}
	hashedUser := user
	hashedUser.Password = "barbar"

	pwHasher.EXPECT().Hash(user.Password).Return(hashedUser.Password, nil)
	repo.EXPECT().Store(hashedUser)

	i := usecases.NewSessionInteractor(
		repo, pwHasher, jwtGen, time.Hour*1,
	)
	err := i.CreateUser(user)

	assert.NoError(t, err)
}
