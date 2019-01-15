package services

import (
	"net/http"
)

// SessionService is responsible for anything user related when they're not logged in such as
// logging in, registering, resetting passwords, requesting new tokens, etc...
type SessionService interface {
	Login(ctx Context) error
	Register(ctx Context) error
}

// NewSessionService initializer
func NewSessionService() SessionService {
	return &sessionService{}
}

type sessionService struct{}

func (s *sessionService) Login(ctx Context) error {
	ctx.NoContent(http.StatusNotImplemented)
	return nil
}

func (s *sessionService) Register(ctx Context) error {
	ctx.NoContent(http.StatusNotImplemented)
	return nil
}
