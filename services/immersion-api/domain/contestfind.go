package domain

import (
	"context"

	"github.com/google/uuid"
)

type ContestFindRepository interface {
	FindContestByID(ctx context.Context, req *ContestFindRequest) (*ContestView, error)
}

type ContestFindRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

type ContestFind struct {
	repo ContestFindRepository
}

func NewContestFind(repo ContestFindRepository) *ContestFind {
	return &ContestFind{repo: repo}
}

func (s *ContestFind) Execute(ctx context.Context, req *ContestFindRequest) (*ContestView, error) {
	req.IncludeDeleted = isAdmin(ctx)

	return s.repo.FindContestByID(ctx, req)
}
