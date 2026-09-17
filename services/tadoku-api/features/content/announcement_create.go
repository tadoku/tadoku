package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/datex"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var ErrInvalidAnnouncement = errors.New("invalid announcement")

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
		!IsValidAnnouncementStyle(p.Style) || !datex.IsValidRange(p.StartsAt, p.EndsAt) {
		return ErrInvalidAnnouncement
	}
	return nil
}

func (s *Service) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) (*Announcement, error) {
	if err := parameters.Validate(); err != nil {
		return nil, err
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
	if err := s.announcements.CreateAnnouncement(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}
