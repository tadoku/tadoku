// Package content owns editorial content and its persistence.
package content

import (
	"context"
	"errors"
	"strconv"
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

type AnnouncementList struct {
	Announcements []Announcement
	TotalSize     int
	NextPageToken string
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

func (s *Service) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*AnnouncementList, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}

	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalSize, err := s.announcements.CountAnnouncements(ctx, namespace)
	if err != nil {
		return nil, err
	}
	announcements, err := s.announcements.ListAnnouncements(ctx, namespace, int32(pageSize), int32(page*pageSize))
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if (page*pageSize)+pageSize < totalSize {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &AnnouncementList{
		Announcements: announcements,
		TotalSize:     totalSize,
		NextPageToken: nextPageToken,
	}, nil
}
