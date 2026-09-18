package pages

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var (
	ErrInvalidPage       = errx.NewInvalidInputError("invalid page")
	ErrPageAlreadyExists = errx.NewConflictError("page already exists")
)

type CreatePageParameters struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	HTML        string
	PublishedAt *time.Time
}

func (p CreatePageParameters) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidPage
	}
	return validatePageFields(p.Namespace, p.Slug, p.Title, p.HTML)
}

func validatePageFields(namespace, slug, title, html string) error {
	if namespace == "" || title == "" || html == "" ||
		utf8.RuneCountInString(slug) <= 1 || slug != strings.ToLower(slug) {
		return ErrInvalidPage
	}
	return nil
}

func (s *Service) CreatePage(ctx context.Context, parameters CreatePageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	now := timex.Now()
	item := &Page{
		ID:          parameters.ID,
		Namespace:   parameters.Namespace,
		Slug:        parameters.Slug,
		Title:       parameters.Title,
		HTML:        parameters.HTML,
		PublishedAt: parameters.PublishedAt,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	return s.pages.CreatePage(ctx, item)
}
