// Package pages owns content pages and their persistence.
package pages

import (
	"strings"
	"time"
	"unicode/utf8"

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
	ErrInvalidSlug       = errx.NewInvalidInputError("slug is required")
	ErrInvalidPage       = errx.NewInvalidInputError("invalid page")
	ErrPageNotFound      = errx.NewNotFoundError("page not found")
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

type PageList struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

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
