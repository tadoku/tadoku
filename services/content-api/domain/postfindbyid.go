package domain

import (
	"context"

	"github.com/google/uuid"
)

type PostFindByIDRepository interface {
	GetPostByID(ctx context.Context, id uuid.UUID, namespace string) (*Post, error)
}

type PostFindByID struct {
	repo PostFindByIDRepository
}

func NewPostFindByID(repo PostFindByIDRepository) *PostFindByID {
	return &PostFindByID{repo: repo}
}

func (s *PostFindByID) Execute(ctx context.Context, id uuid.UUID, namespace string) (*Post, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.GetPostByID(ctx, id, namespace)
}
