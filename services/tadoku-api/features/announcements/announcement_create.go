package announcements

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var ErrInvalidAnnouncement = errx.NewInvalidInputError("invalid announcement")

type CreateAnnouncementParameters struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
}

func (p CreateAnnouncementParameters) Validate() error {
	if p.ID == uuid.Nil || p.Namespace == "" || p.Title == "" || p.Content == "" ||
		!isValidAnnouncementStyle(p.Style) || !isValidAnnouncementHref(p.Href) ||
		!timex.IsValidRange(p.StartsAt, p.EndsAt) {
		return ErrInvalidAnnouncement
	}
	return nil
}

func (s *Service) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	now := timex.Now()
	item := &Announcement{
		ID:        parameters.ID,
		Namespace: parameters.Namespace,
		Title:     parameters.Title,
		Content:   parameters.Content,
		Style:     parameters.Style,
		Href:      parameters.Href,
		StartsAt:  parameters.StartsAt,
		EndsAt:    parameters.EndsAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.announcements.CreateAnnouncement(ctx, item)
}
