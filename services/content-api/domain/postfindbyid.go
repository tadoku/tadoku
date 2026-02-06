package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PostFindByIDRepository interface {
	GetPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
}

type PostFindByID struct {
	repo PostFindByIDRepository
}

func NewPostFindByID(repo PostFindByIDRepository) *PostFindByID {
	return &PostFindByID{repo: repo}
}

func (s *PostFindByID) Execute(ctx context.Context, id uuid.UUID) (*Post, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetPostByID(ctx, id)
}
