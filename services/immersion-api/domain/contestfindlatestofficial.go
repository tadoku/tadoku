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
	contest, err := s.repo.ContestFindLatestOfficial(ctx)
	if err != nil {
		return nil, err
	}
	if err := hydrateContestActivities(contest, true); err != nil {
		return nil, err
	}
	return contest, nil
}
