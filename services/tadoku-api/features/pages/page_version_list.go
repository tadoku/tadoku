package pages

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) ListPageVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PageVersion, error) {
	return s.pages.ListPageVersions(ctx, namespace, id)
}
