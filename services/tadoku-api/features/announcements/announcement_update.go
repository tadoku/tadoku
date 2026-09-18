package announcements

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type UpdateAnnouncementParameters struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
}

func (p UpdateAnnouncementParameters) Validate() error {
	if p.Namespace == "" || p.Title == "" || p.Content == "" ||
		!IsValidAnnouncementStyle(p.Style) || !isValidAnnouncementHref(p.Href) ||
		!timex.IsValidRange(p.StartsAt, p.EndsAt) {
		return ErrInvalidAnnouncement
	}
	return nil
}

func (s *Service) UpdateAnnouncement(ctx context.Context, parameters UpdateAnnouncementParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	announcement, err := s.announcements.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
	if err != nil {
		return err
	}

	announcement.Title = parameters.Title
	announcement.Content = parameters.Content
	announcement.Style = parameters.Style
	announcement.Href = parameters.Href
	announcement.StartsAt = parameters.StartsAt
	announcement.EndsAt = parameters.EndsAt
	announcement.UpdatedAt = timex.Now()

	return s.announcements.UpdateAnnouncement(ctx, announcement)
}
