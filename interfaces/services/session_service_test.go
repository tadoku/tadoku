package services_test

import (
	"net/http"
	"testing"
	"time"

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

	s := services.NewSessionService(i, "session_cookie")
	err := s.Register(ctx)

	assert.NoError(t, err)
}

func TestSessionService_Login(t *testing.T) {
	expiresAt := time.Now().Unix()
	user := &domain.User{
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Password:    "foobar",
	}
	cookieName := "session_cookie"
	token := "foobar"

	b := &services.SessionLoginBody{
		Email:    "foo@bar.com",
		Password: "foobar",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().JSON(200, map[string]interface{}{
		"expiresAt": expiresAt,
		"user":      *user,
	})
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *b)
	ctx.EXPECT().SetCookie(gomock.Any()).Do(func(cookie *http.Cookie) {
		assert.Equal(t, cookieName, cookie.Name)
		assert.Equal(t, token, cookie.Value)
		assert.Equal(t, expiresAt, cookie.Expires.Unix())
		assert.True(t, cookie.Secure)
		assert.True(t, cookie.HttpOnly)
	})

	i := usecases.NewMockSessionInteractor(ctrl)
	i.EXPECT().CreateSession(b.Email, b.Password).Return(*user, token, expiresAt, nil)

	s := services.NewSessionService(i, cookieName)
	err := s.Login(ctx)

	assert.NoError(t, err)
}

func TestSessionService_Refresh(t *testing.T) {
	user := &domain.User{
		ID:          1,
		Email:       "foo@bar.com",
		DisplayName: "John Doe",
		Role:        domain.RoleUser,
	}
	token := "foobar"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().JSON(200, map[string]interface{}{
		"token": token,
		"user":  *user,
	})
	ctx.EXPECT().User().Return(user, nil)

	i := usecases.NewMockSessionInteractor(ctrl)
	i.EXPECT().RefreshSession(*user).Return(*user, token, nil)

	s := services.NewSessionService(i, "session_cookie")
	err := s.Refresh(ctx)

	assert.NoError(t, err)
}
