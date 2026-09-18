package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (s *Service) FindPostBySlug(ctx context.Context, namespace, slug string) (*Post, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	if slug == "" {
		return nil, errx.NewInvalidInputError("slug is required")
	}

	post, err := s.posts.FindPostBySlug(ctx, namespace, slug)
	if err != nil {
		return nil, err
	}
	if post.PublishedAt == nil || post.PublishedAt.After(timex.Now()) {
		return nil, ErrPostNotFound
	}

	return post, nil
}

func (s *Service) FindPostByID(ctx context.Context, namespace string, id uuid.UUID) (*Post, error) {
	return s.posts.FindPostByID(ctx, namespace, id)
}
