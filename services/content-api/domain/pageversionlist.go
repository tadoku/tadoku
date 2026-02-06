package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PageVersionListRepository interface {
	ListPageVersions(ctx context.Context, pageID uuid.UUID) ([]PageVersion, error)
}

type PageVersionList struct {
	repo PageVersionListRepository
}

func NewPageVersionList(repo PageVersionListRepository) *PageVersionList {
	return &PageVersionList{
		repo: repo,
	}
}

func (s *PageVersionList) Execute(ctx context.Context, pageID uuid.UUID) ([]PageVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.ListPageVersions(ctx, pageID)
}
