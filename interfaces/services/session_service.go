package services

import (
	"net/http"

	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// SessionService is responsible for anything user related when they're not logged in such as
// logging in, registering, resetting passwords, requesting new tokens, etc...
type SessionService interface {
	Login(ctx Context) error
	Register(ctx Context) error
}

// NewSessionService initializer
func NewSessionService(sessionInteractor usecases.SessionInteractor) SessionService {
	return &sessionService{
		SessionInteractor: sessionInteractor,
	}
}

type sessionService struct {
	SessionInteractor usecases.SessionInteractor
}

func (s *sessionService) Login(ctx Context) error {
	ctx.NoContent(http.StatusNotImplemented)
	return nil
}

func (s *sessionService) Register(ctx Context) error {
	user := &domain.User{}
	err := ctx.Bind(user)
	if err != nil {
		return fail.Wrap(err)
	}

	s.SessionInteractor.CreateUser(*user)

	return ctx.NoContent(http.StatusCreated)
}
