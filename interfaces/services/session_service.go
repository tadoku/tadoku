package services

import (
	"net/http"
	"time"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// SessionService is responsible for anything user related when they're not logged in such as
// logging in, registering, resetting passwords, requesting new tokens, etc...
type SessionService interface {
	Login(ctx Context) error
	Refresh(ctx Context) error
}

// NewSessionService initializer
func NewSessionService(
	sessionInteractor usecases.SessionInteractor,
	sessionCookieName string,
) SessionService {
	return &sessionService{
		SessionInteractor: sessionInteractor,
		sessionCookieName: sessionCookieName,
	}
}

type sessionService struct {
	SessionInteractor usecases.SessionInteractor
	sessionCookieName string
}

// SessionLoginBody is the data that's needed to log in
type SessionLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *sessionService) Login(ctx Context) error {
	b := &SessionLoginBody{}
	err := ctx.Bind(b)
	if err != nil {
		return domain.WrapError(err)
	}

	user, token, expiresAt, err := s.SessionInteractor.CreateSession(b.Email, b.Password)
	if err != nil {
		ctx.NoContent(http.StatusUnauthorized)
		return domain.WrapError(err)
	}

	res := map[string]interface{}{
		"expiresAt": expiresAt,
		"user":      user,
	}

	sessionCookie := &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    token,
		Expires:  time.Unix(expiresAt, 0),
		Secure:   true,
		HttpOnly: true,
	}

	ctx.SetCookie(sessionCookie)

	return ctx.JSON(http.StatusOK, res)
}

func (s *sessionService) Refresh(ctx Context) error {
	sessionUser, err := ctx.User()
	if err != nil {
		return domain.WrapError(err)
	}

	user, token, _, err := s.SessionInteractor.RefreshSession(*sessionUser)
	if err != nil {
		ctx.NoContent(http.StatusUnauthorized)
		return domain.WrapError(err)
	}

	res := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	return ctx.JSON(http.StatusOK, res)
}
