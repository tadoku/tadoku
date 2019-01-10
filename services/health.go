package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/domain/services"
)

func NewHealthService() services.HealthService {
	return &healthService{}
}

type healthService struct{}

func (s *healthService) Ping(ctx domain.Context) error {
	ctx.String(http.StatusOK, "pong")
	return nil
}
