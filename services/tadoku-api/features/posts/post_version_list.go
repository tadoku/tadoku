package posts

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PostVersion, error) {
	return s.posts.ListPostVersions(ctx, namespace, id)
}
