package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PostVersion struct {
	ID        uuid.UUID
	Version   int
	Title     string
	Content   string
	CreatedAt time.Time
}

type PostVersionGetRepository interface {
	GetPostVersion(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*PostVersion, error)
}

type PostVersionGet struct {
	repo PostVersionGetRepository
}

func NewPostVersionGet(repo PostVersionGetRepository) *PostVersionGet {
	return &PostVersionGet{
		repo: repo,
	}
}

func (s *PostVersionGet) Execute(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*PostVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetPostVersion(ctx, postID, contentID)
}
