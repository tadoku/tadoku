package domain

import (
	"context"

	"github.com/google/uuid"
)

type PageDeleteRepository interface {
	DeletePage(ctx context.Context, id uuid.UUID, namespace string) error
}

type PageDelete struct {
	repo PageDeleteRepository
}

func NewPageDelete(repo PageDeleteRepository) *PageDelete {
	return &PageDelete{
		repo: repo,
	}
}

func (s *PageDelete) Execute(ctx context.Context, id uuid.UUID, namespace string) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}

	return s.repo.DeletePage(ctx, id, namespace)
}
