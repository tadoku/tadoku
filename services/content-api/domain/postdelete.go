package domain

import (
	"context"

	"github.com/google/uuid"
)

type PostDeleteRepository interface {
	DeletePost(ctx context.Context, id uuid.UUID, namespace string) error
}

type PostDelete struct {
	repo PostDeleteRepository
}

func NewPostDelete(repo PostDeleteRepository) *PostDelete {
	return &PostDelete{
		repo: repo,
	}
}

func (s *PostDelete) Execute(ctx context.Context, id uuid.UUID, namespace string) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}

	return s.repo.DeletePost(ctx, id, namespace)
}
