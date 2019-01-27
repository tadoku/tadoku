package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestSessionService_Register(t *testing.T) {
	user := &domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *user)

	user.Role = domain.RoleUser
	user.Preferences = &domain.Preferences{}

	i := usecases.NewMockSessionInteractor(ctrl)
	i.EXPECT().CreateUser(*user).Return(nil)

	s := services.NewSessionService(i)
	err := s.Register(ctx)

	assert.NoError(t, err)
}
