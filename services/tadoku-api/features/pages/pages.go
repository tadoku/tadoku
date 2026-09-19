// Package pages owns content pages and their persistence.
package pages

import (
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Page struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	HTML        string
	PublishedAt *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

type PageVersion struct {
	ID        uuid.UUID
	Version   int
	Title     string
	HTML      string
	CreatedAt time.Time
}

var (
	ErrInvalidNamespace  = errx.NewInvalidInputError("namespace is required")
	ErrInvalidPagination = errx.NewInvalidInputError("invalid pagination")
	ErrPageNotFound      = errx.NewNotFoundError("page not found")
)

type Service struct {
	pages *PagesRepository
}

func NewService(pages *PagesRepository) *Service {
	return &Service{pages: pages}
}
