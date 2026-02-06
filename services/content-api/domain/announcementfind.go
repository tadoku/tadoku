package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type AnnouncementFindByIDRepository interface {
	GetAnnouncementByID(ctx context.Context, id uuid.UUID, namespace string) (*Announcement, error)
}

type AnnouncementFindByID struct {
	repo AnnouncementFindByIDRepository
}

func NewAnnouncementFindByID(repo AnnouncementFindByIDRepository) *AnnouncementFindByID {
	return &AnnouncementFindByID{
		repo: repo,
	}
}

func (s *AnnouncementFindByID) Execute(ctx context.Context, id uuid.UUID, namespace string) (*Announcement, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetAnnouncementByID(ctx, id, namespace)
}
