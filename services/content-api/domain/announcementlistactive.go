package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AnnouncementListActiveRepository interface {
	ListActiveAnnouncements(ctx context.Context, namespace string, now time.Time) ([]Announcement, error)
}

type AnnouncementListActiveRequest struct {
	Namespace string `validate:"required"`
}

type AnnouncementListActiveResponse struct {
	Announcements []Announcement
}

type AnnouncementListActive struct {
	repo     AnnouncementListActiveRepository
	clock    commondomain.Clock
	validate *validator.Validate
}

func NewAnnouncementListActive(repo AnnouncementListActiveRepository, clock commondomain.Clock) *AnnouncementListActive {
	return &AnnouncementListActive{
		repo:     repo,
		clock:    clock,
		validate: validator.New(),
	}
}

func (s *AnnouncementListActive) Execute(ctx context.Context, req *AnnouncementListActiveRequest) (*AnnouncementListActiveResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	announcements, err := s.repo.ListActiveAnnouncements(ctx, req.Namespace, s.clock.Now())
	if err != nil {
		return nil, err
	}

	return &AnnouncementListActiveResponse{Announcements: announcements}, nil
}
