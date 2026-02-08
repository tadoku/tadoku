package domain

import (
	"context"

	"github.com/google/uuid"
)

type AnnouncementDeleteRepository interface {
	DeleteAnnouncement(ctx context.Context, id uuid.UUID, namespace string) error
}

type AnnouncementDelete struct {
	repo AnnouncementDeleteRepository
}

func NewAnnouncementDelete(repo AnnouncementDeleteRepository) *AnnouncementDelete {
	return &AnnouncementDelete{
		repo: repo,
	}
}

func (s *AnnouncementDelete) Execute(ctx context.Context, id uuid.UUID, namespace string) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}

	return s.repo.DeleteAnnouncement(ctx, id, namespace)
}
