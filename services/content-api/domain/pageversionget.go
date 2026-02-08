package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PageVersion struct {
	ID        uuid.UUID
	Version   int
	Title     string
	HTML      string
	CreatedAt time.Time
}

type PageVersionGetRepository interface {
	GetPageVersion(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*PageVersion, error)
}

type PageVersionGet struct {
	repo PageVersionGetRepository
}

func NewPageVersionGet(repo PageVersionGetRepository) *PageVersionGet {
	return &PageVersionGet{
		repo: repo,
	}
}

func (s *PageVersionGet) Execute(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*PageVersion, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	return s.repo.GetPageVersion(ctx, pageID, contentID)
}
