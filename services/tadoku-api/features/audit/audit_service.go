package audit

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	event.recordedAt = timex.Now()
	return s.repository.Create(ctx, event)
}
