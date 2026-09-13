// Package content owns editorial content and its persistence.
package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

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

var ErrInvalidNamespace = errors.New("namespace is required")

type Service struct {
	announcements *AnnouncementsRepository
}

func NewService(announcements *AnnouncementsRepository) *Service {
	return &Service{
		announcements: announcements,
	}
}

func (s *Service) ListActiveAnnouncements(ctx context.Context, namespace string) ([]Announcement, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}

	// Publication policy belongs here; persistence only applies these inputs.
	return s.announcements.ListActiveAnnouncements(ctx, namespace, timex.Now(), 10)
}
