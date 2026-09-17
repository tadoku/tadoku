package content

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
	return validateAnnouncementFields(p.Namespace, p.Title, p.Content, p.Style, p.StartsAt, p.EndsAt)
}

func (s *Service) UpdateAnnouncement(ctx context.Context, parameters UpdateAnnouncementParameters) (*Announcement, error) {
	if err := parameters.Validate(); err != nil {
		return nil, err
	}

	announcement, err := s.announcements.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
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
