package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AnnouncementUpdateRepository interface {
	GetAnnouncementByID(ctx context.Context, id uuid.UUID, namespace string) (*Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement *Announcement) error
}

type AnnouncementUpdateRequest struct {
	Namespace string `validate:"required"`
	Title     string `validate:"required"`
	Content   string `validate:"required"`
	Style     string `validate:"required,oneof=success warning error info"`
	Href      *string
	StartsAt  time.Time `validate:"required"`
	EndsAt    time.Time `validate:"required,gtfield=StartsAt"`
}

type AnnouncementUpdateResponse struct {
	Announcement *Announcement
}

type AnnouncementUpdate struct {
	repo     AnnouncementUpdateRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewAnnouncementUpdate(repo AnnouncementUpdateRepository, clock commondomain.Clock) *AnnouncementUpdate {
	return &AnnouncementUpdate{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

func (s *AnnouncementUpdate) Execute(ctx context.Context, id uuid.UUID, req *AnnouncementUpdateRequest) (*AnnouncementUpdateResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAnnouncement, err)
	}

	announcement, err := s.repo.GetAnnouncementByID(ctx, id, req.Namespace)
	if err != nil {
		return nil, err
	}

	announcement.Namespace = req.Namespace
	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Style = req.Style
	announcement.Href = req.Href
	announcement.StartsAt = req.StartsAt
	announcement.EndsAt = req.EndsAt
	announcement.UpdatedAt = s.clock.Now()

	if err := s.repo.UpdateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}

	return &AnnouncementUpdateResponse{Announcement: announcement}, nil
}
