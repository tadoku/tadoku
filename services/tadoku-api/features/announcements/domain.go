// Package announcements owns announcements and their persistence.
package announcements

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const announcementHrefMaxLength = 2048

type Announcement struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementList struct {
	Announcements []Announcement
	TotalSize     int
	NextPageToken string
}

func isValidAnnouncementStyle(style string) bool {
	switch style {
	case "success", "warning", "error", "info":
		return true
	default:
		return false
	}
}

func isValidAnnouncementHref(href string) bool {
	parsed, err := url.Parse(href)
	if err != nil {
		return false
	}
	if parsed.Scheme == "http" || parsed.Scheme == "https" {
		return true
	}
	return parsed.Scheme == "" && strings.HasPrefix(href, "/") &&
		!strings.HasPrefix(href, "//") && !strings.HasPrefix(href, `/\`)
}

var (
	ErrAnnouncementNotFound      = errx.NewNotFoundError("announcement not found")
	ErrAnnouncementAlreadyExists = errx.NewConflictError("announcement already exists")
)

type announcementParameters struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
}

type CreateAnnouncementParameters = announcementParameters
type UpdateAnnouncementParameters = announcementParameters

func (p announcementParameters) Validate() error {
	if p.ID == uuid.Nil {
		return errx.NewInvalidInputError("id is required")
	}
	if p.Namespace == "" {
		return errx.NewInvalidInputError("namespace is required")
	}
	if p.Title == "" {
		return errx.NewInvalidInputError("title is required")
	}
	if p.Content == "" {
		return errx.NewInvalidInputError("content is required")
	}
	if !isValidAnnouncementStyle(p.Style) {
		return errx.NewInvalidInputError("style must be one of success, warning, error or info")
	}
	if p.Href != nil && utf8.RuneCountInString(*p.Href) > announcementHrefMaxLength {
		return errx.NewInvalidInputError("href must be at most 2048 characters")
	}
	if p.Href != nil && *p.Href != "" && !isValidAnnouncementHref(*p.Href) {
		return errx.NewInvalidInputError("href must be an http(s) URL or a root-relative path")
	}
	if p.StartsAt.IsZero() {
		return errx.NewInvalidInputError("starts_at is required")
	}
	if p.EndsAt.IsZero() {
		return errx.NewInvalidInputError("ends_at is required")
	}
	if !p.StartsAt.Before(p.EndsAt) {
		return errx.NewInvalidInputError("ends_at must be after starts_at")
	}
	return nil
}
