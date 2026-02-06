package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PageFindByIDRepository interface {
	GetPageByID(ctx context.Context, id uuid.UUID) (*Page, error)
}

type PageFindByID struct {
	repo PageFindByIDRepository
}

func NewPageFindByID(repo PageFindByIDRepository) *PageFindByID {
	return &PageFindByID{repo: repo}
}

func (s *PageFindByID) Execute(ctx context.Context, id uuid.UUID) (*Page, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetPageByID(ctx, id)
}
