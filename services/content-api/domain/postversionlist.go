package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.ListPostVersions(ctx, postID)
}
