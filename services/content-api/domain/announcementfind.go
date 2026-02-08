package domain

import (
	"context"

	"github.com/google/uuid"
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
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.GetAnnouncementByID(ctx, id, namespace)
}
