package domain

import (
	"context"

	"github.com/google/uuid"
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
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.ListPageVersions(ctx, pageID)
}
