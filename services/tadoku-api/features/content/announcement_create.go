package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var ErrInvalidAnnouncement = errors.New("invalid announcement")

type CreateAnnouncementRequest struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
}

func (s *Service) CreateAnnouncement(ctx context.Context, request CreateAnnouncementRequest) (*Announcement, error) {
	if request.ID == uuid.Nil || request.Namespace == "" || request.Title == "" || request.Content == "" ||
		request.StartsAt.IsZero() || request.EndsAt.IsZero() || !request.EndsAt.After(request.StartsAt) {
		return nil, ErrInvalidAnnouncement
	}
	switch request.Style {
	case "success", "warning", "error", "info":
	default:
		return nil, ErrInvalidAnnouncement
	}

	now := timex.Now()
	item := &Announcement{
		ID:        request.ID,
		Namespace: request.Namespace,
		Title:     request.Title,
		Content:   request.Content,
		Style:     request.Style,
		Href:      request.Href,
		StartsAt:  request.StartsAt,
		EndsAt:    request.EndsAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.announcements.CreateAnnouncement(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}
