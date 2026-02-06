package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return ErrForbidden
	}

	return s.repo.DeletePost(ctx, id, namespace)
}
