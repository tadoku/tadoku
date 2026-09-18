package pages

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (s *Service) DeletePage(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.pages.DeletePage(ctx, namespace, id, timex.Now())
}
