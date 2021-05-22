package services

import (
	"net/http"
)

// HealthService is responsible for metrics about the health of the service
type HealthService interface {
	// Ping is only used to see if the service is online, it doesn't do any health checks
	Ping(ctx Context) error
}

// NewHealthService initializer
func NewHealthService() HealthService {
	return &healthService{}
}

type healthService struct{}

func (s *healthService) Ping(ctx Context) error {
	ctx.String(http.StatusOK, "pong")
	return nil
}
