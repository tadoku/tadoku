package services

import (
	"github.com/tadoku/api/domain"
)

type HealthService interface {
	Ping(ctx domain.Context) error
}
