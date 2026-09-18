package pages

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (s *Service) FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	if slug == "" {
		return nil, errx.NewInvalidInputError("slug is required")
	}

	page, err := s.pages.FindPageBySlug(ctx, namespace, slug)
	if err != nil {
		return nil, err
	}
	if page.PublishedAt == nil || page.PublishedAt.After(timex.Now()) {
		return nil, ErrPageNotFound
	}

	return page, nil
}

func (s *Service) FindPageByID(ctx context.Context, namespace string, id uuid.UUID) (*Page, error) {
	return s.pages.FindPageByID(ctx, namespace, id)
}
