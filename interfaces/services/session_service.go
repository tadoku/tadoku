package services

import (
	"net/http"
)

type SessionService interface {
	Login(ctx Context) error
	Register(ctx Context) error
}

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
