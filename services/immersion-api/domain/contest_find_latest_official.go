package domain

import (
	"context"
)

type ContestFindLatestOfficialRepository interface {
	ContestFindLatestOfficial(ctx context.Context) (*ContestView, error)
}

type ContestFindLatestOfficial struct {
	repo ContestFindLatestOfficialRepository
}

func NewContestFindLatestOfficial(repo ContestFindLatestOfficialRepository) *ContestFindLatestOfficial {
	return &ContestFindLatestOfficial{repo: repo}
}

func (s *ContestFindLatestOfficial) Execute(ctx context.Context) (*ContestView, error) {
	return s.repo.ContestFindLatestOfficial(ctx)
}
