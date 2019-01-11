package services

import (
	"net/http"
)

type HealthService interface {
	Ping(ctx Context) error
}

func NewHealthService() HealthService {
	return &healthService{}
}

type healthService struct{}

func (s *healthService) Ping(ctx Context) error {
	ctx.String(http.StatusOK, "pong")
	return nil
}
