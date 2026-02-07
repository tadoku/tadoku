package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AnnouncementCreateRepository interface {
	CreateAnnouncement(ctx context.Context, announcement *Announcement) error
}

type AnnouncementCreateRequest struct {
	ID        uuid.UUID `validate:"required"`
	Namespace string    `validate:"required"`
	Title     string    `validate:"required"`
	Content   string    `validate:"required"`
	Style     string    `validate:"required,oneof=success warning error info"`
	Href      *string
	StartsAt  time.Time `validate:"required"`
	EndsAt    time.Time `validate:"required,gtfield=StartsAt"`
}

type AnnouncementCreateResponse struct {
	Announcement *Announcement
}

type AnnouncementCreate struct {
	repo     AnnouncementCreateRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewAnnouncementCreate(repo AnnouncementCreateRepository, clock commondomain.Clock) *AnnouncementCreate {
	return &AnnouncementCreate{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

func (s *AnnouncementCreate) Execute(ctx context.Context, req *AnnouncementCreateRequest) (*AnnouncementCreateResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAnnouncement, err)
	}

	now := s.clock.Now()
	announcement := &Announcement{
		ID:        req.ID,
		Namespace: req.Namespace,
		Title:     req.Title,
		Content:   req.Content,
		Style:     req.Style,
		Href:      req.Href,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}

	return &AnnouncementCreateResponse{Announcement: announcement}, nil
}
