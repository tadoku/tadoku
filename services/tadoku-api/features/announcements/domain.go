// Package announcements owns announcements and their persistence.
package announcements

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
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

func isValidAnnouncementHref(href *string) bool {
	if href == nil || *href == "" {
		return true
	}
	if utf8.RuneCountInString(*href) > announcementHrefMaxLength {
		return false
	}

	parsed, err := url.Parse(*href)
	if err != nil {
		return false
	}
	if parsed.Scheme == "http" || parsed.Scheme == "https" {
		return true
	}
	return parsed.Scheme == "" && strings.HasPrefix(*href, "/") &&
		!strings.HasPrefix(*href, "//") && !strings.HasPrefix(*href, `/\`)
}

var (
	ErrInvalidNamespace          = errx.NewInvalidInputError("namespace is required")
	ErrInvalidPagination         = errx.NewInvalidInputError("invalid pagination")
	ErrInvalidAnnouncement       = errx.NewInvalidInputError("invalid announcement")
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
		return fmt.Errorf("%w: id is nil", ErrInvalidAnnouncement)
	}
	if p.Namespace == "" {
		return fmt.Errorf("%w: namespace is required", ErrInvalidAnnouncement)
	}
	if p.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidAnnouncement)
	}
	if p.Content == "" {
		return fmt.Errorf("%w: content is required", ErrInvalidAnnouncement)
	}
	if !isValidAnnouncementStyle(p.Style) {
		return fmt.Errorf("%w: style is invalid", ErrInvalidAnnouncement)
	}
	if !isValidAnnouncementHref(p.Href) {
		return fmt.Errorf("%w: href is invalid", ErrInvalidAnnouncement)
	}
	if !timex.IsValidRange(p.StartsAt, p.EndsAt) {
		return fmt.Errorf("%w: date range is invalid", ErrInvalidAnnouncement)
	}
	return nil
}
