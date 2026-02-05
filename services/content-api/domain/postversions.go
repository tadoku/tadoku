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

type PostVersionListRepository interface {
	ListPostVersions(ctx context.Context, postID uuid.UUID) ([]PostVersion, error)
	GetPostVersion(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*PostVersion, error)
}

type PostVersionList struct {
	repo PostVersionListRepository
}

func NewPostVersionList(repo PostVersionListRepository) *PostVersionList {
	return &PostVersionList{
		repo: repo,
	}
}

func (s *PostVersionList) List(ctx context.Context, postID uuid.UUID) ([]PostVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.ListPostVersions(ctx, postID)
}

func (s *PostVersionList) Get(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*PostVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetPostVersion(ctx, postID, contentID)
}
