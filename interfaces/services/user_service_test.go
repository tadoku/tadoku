package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestUserService_UpdatePassword(t *testing.T) {
	user := &domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}

	b := &services.UserUpdatePasswordBody{
		CurrentPassword: "foobar",
		NewPassword:     "barbar",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(200)
	ctx.EXPECT().User().Return(user, nil)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *b)

	i := usecases.NewMockUserInteractor(ctrl)
	i.EXPECT().UpdatePassword(user.Email, b.CurrentPassword, b.NewPassword).Return(nil)

	s := services.NewUserService(i)
	err := s.UpdatePassword(ctx)

	assert.NoError(t, err)
}

func TestUserService_UpdateProfile(t *testing.T) {
	user := domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}

	b := &services.UserUpdateProfileBody{
		DisplayName: "foobar",
	}

	updatedUser := user
	updatedUser.DisplayName = b.DisplayName

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(200)
	ctx.EXPECT().User().Return(&user, nil)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *b)

	i := usecases.NewMockUserInteractor(ctrl)
	i.EXPECT().UpdateProfile(updatedUser).Return(nil)

	s := services.NewUserService(i)
	err := s.UpdateProfile(ctx)

	assert.NoError(t, err)
}
