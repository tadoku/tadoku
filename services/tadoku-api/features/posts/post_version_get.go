package posts

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*PostVersion, error) {
	return s.posts.GetPostVersion(ctx, namespace, postID, contentID)
}
