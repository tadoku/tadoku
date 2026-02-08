package domain

import (
	"context"

	"github.com/google/uuid"
)

type PageFindByIDRepository interface {
	GetPageByID(ctx context.Context, id uuid.UUID, namespace string) (*Page, error)
}

type PageFindByID struct {
	repo PageFindByIDRepository
}

func NewPageFindByID(repo PageFindByIDRepository) *PageFindByID {
	return &PageFindByID{repo: repo}
}

func (s *PageFindByID) Execute(ctx context.Context, id uuid.UUID, namespace string) (*Page, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.GetPageByID(ctx, id, namespace)
}
