package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return ErrForbidden
	}

	return s.repo.DeletePage(ctx, id, namespace)
}
