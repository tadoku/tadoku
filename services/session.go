package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/domain/services"
)

func NewSessionService() services.SessionService {
	return &sessionService{}
}

type sessionService struct{}

func (s *sessionService) Login(ctx domain.Context) error {
	ctx.NoContent(http.StatusNotImplemented)
	return nil
}

func (s *sessionService) Register(ctx domain.Context) error {
	ctx.NoContent(http.StatusNotImplemented)
	return nil
}
