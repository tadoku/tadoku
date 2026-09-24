package announcements

import (
	"context"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	announcements *AnnouncementsRepository
}

func NewService(announcements *AnnouncementsRepository) *Service {
	return &Service{
		announcements: announcements,
	}
}

func (s *Service) FindAnnouncementByID(ctx context.Context, namespace string, id uuid.UUID) (*Announcement, error) {
	return s.announcements.FindAnnouncementByID(ctx, namespace, id)
}

func (s *Service) ListActiveAnnouncements(ctx context.Context, namespace string) ([]Announcement, error) {
	if namespace == "" {
		return nil, errx.NewInvalidInputError("namespace is required")
	}

	// Publication policy belongs here; persistence only applies these inputs.
	return s.announcements.ListActiveAnnouncements(ctx, namespace, timex.Now(), 10)
}

func (s *Service) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*AnnouncementList, error) {
	if namespace == "" {
		return nil, errx.NewInvalidInputError("namespace is required")
	}

	if pageSize < 0 {
		return nil, errx.NewInvalidInputError("page_size must not be negative")
	}
	if page < 0 {
		return nil, errx.NewInvalidInputError("page must not be negative")
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int64(page)
	if page > math.MaxInt64/pageSize {
		offset = math.MaxInt64
	} else {
		offset *= int64(pageSize)
	}

	announcements, totalSize, err := s.announcements.ListAnnouncements(ctx, namespace, int32(pageSize), offset)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if int64(pageSize) < int64(totalSize)-offset {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &AnnouncementList{
		Announcements: announcements,
		TotalSize:     totalSize,
		NextPageToken: nextPageToken,
	}, nil
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

func (s *Service) DeleteAnnouncement(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.announcements.DeleteAnnouncement(ctx, namespace, id, timex.Now())
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
