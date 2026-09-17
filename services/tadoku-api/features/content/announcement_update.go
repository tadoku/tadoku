package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/datex"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type UpdateAnnouncementParameters struct {
	Title    string
	Content  string
	Style    string
	Href     *string
	StartsAt time.Time
	EndsAt   time.Time
}

func (p UpdateAnnouncementParameters) Validate(namespace string) error {
	if namespace == "" || p.Title == "" || p.Content == "" ||
		!IsValidAnnouncementStyle(p.Style) || !datex.IsValidRange(p.StartsAt, p.EndsAt) {
		return ErrInvalidAnnouncement
	}
	return nil
}

func (s *Service) UpdateAnnouncement(ctx context.Context, namespace string, id uuid.UUID, parameters UpdateAnnouncementParameters) (*Announcement, error) {
	if err := parameters.Validate(namespace); err != nil {
		return nil, err
	}

	announcement, err := s.announcements.FindAnnouncementByID(ctx, namespace, id)
	if err != nil {
		return nil, err
	}

	announcement.Title = parameters.Title
	announcement.Content = parameters.Content
	announcement.Style = parameters.Style
	announcement.Href = parameters.Href
	announcement.StartsAt = parameters.StartsAt
	announcement.EndsAt = parameters.EndsAt
	announcement.UpdatedAt = timex.Now()

	if err := s.announcements.UpdateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}
	return announcement, nil
}
