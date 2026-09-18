package posts

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (s *Service) DeletePost(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.posts.DeletePost(ctx, namespace, id, timex.Now())
}
