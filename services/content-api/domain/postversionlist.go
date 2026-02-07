package domain

import (
	"context"

	"github.com/google/uuid"
)

type PostVersionListRepository interface {
	ListPostVersions(ctx context.Context, postID uuid.UUID) ([]PostVersion, error)
}

type PostVersionList struct {
	repo PostVersionListRepository
}

func NewPostVersionList(repo PostVersionListRepository) *PostVersionList {
	return &PostVersionList{
		repo: repo,
	}
}

func (s *PostVersionList) Execute(ctx context.Context, postID uuid.UUID) ([]PostVersion, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.ListPostVersions(ctx, postID)
}
