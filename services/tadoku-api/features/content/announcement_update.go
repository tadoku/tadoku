package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type AnnouncementUpdateRequest struct {
	Title    string
	Content  string
	Style    string
	Href     *string
	StartsAt time.Time
	EndsAt   time.Time
}

func (req AnnouncementUpdateRequest) Validate(namespace string) error {
	if namespace == "" || req.Title == "" || req.Content == "" ||
		req.StartsAt.IsZero() || req.EndsAt.IsZero() || !req.EndsAt.After(req.StartsAt) {
		return ErrInvalidAnnouncement
	}
	switch req.Style {
	case "success", "warning", "error", "info":
	default:
		return ErrInvalidAnnouncement
	}
	return nil
}

func (s *Service) UpdateAnnouncement(ctx context.Context, namespace string, id uuid.UUID, req AnnouncementUpdateRequest) (*Announcement, error) {
	if err := req.Validate(namespace); err != nil {
		return nil, err
	}

	announcement, err := s.announcements.FindAnnouncementByID(ctx, namespace, id)
	if err != nil {
		return nil, err
	}

	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Style = req.Style
	announcement.Href = req.Href
	announcement.StartsAt = req.StartsAt
	announcement.EndsAt = req.EndsAt
	announcement.UpdatedAt = timex.Now()

	if err := s.announcements.UpdateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}
	return announcement, nil
}
