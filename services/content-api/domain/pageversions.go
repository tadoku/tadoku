package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PageVersion struct {
	ID        uuid.UUID
	Version   int
	Title     string
	HTML      string
	CreatedAt time.Time
}

type PageVersionListRepository interface {
	ListPageVersions(ctx context.Context, pageID uuid.UUID) ([]PageVersion, error)
	GetPageVersion(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*PageVersion, error)
}

type PageVersionList struct {
	repo PageVersionListRepository
}

func NewPageVersionList(repo PageVersionListRepository) *PageVersionList {
	return &PageVersionList{
		repo: repo,
	}
}

func (s *PageVersionList) List(ctx context.Context, pageID uuid.UUID) ([]PageVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.ListPageVersions(ctx, pageID)
}

func (s *PageVersionList) Get(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*PageVersion, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	return s.repo.GetPageVersion(ctx, pageID, contentID)
}
