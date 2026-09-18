package pages

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type UpdatePageParameters struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	HTML        string
	PublishedAt *time.Time
}

func (p UpdatePageParameters) Validate() error {
	return validatePageFields(p.Namespace, p.Slug, p.Title, p.HTML)
}

func (s *Service) UpdatePage(ctx context.Context, parameters UpdatePageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	item, err := s.pages.FindPageByID(ctx, parameters.Namespace, parameters.ID)
	if err != nil {
		return err
	}

	contentChanged := item.Title != parameters.Title || item.HTML != parameters.HTML
	item.Slug = parameters.Slug
	item.Title = parameters.Title
	item.HTML = parameters.HTML
	item.PublishedAt = parameters.PublishedAt
	now := timex.Now()
	item.UpdatedAt = &now

	return s.pages.UpdatePage(ctx, item, contentChanged)
}
