package domain

import (
	"context"
)

type ContestConfigurationOptionsRepository interface {
	FetchContestConfigurationOptions(ctx context.Context) (*ContestConfigurationOptionsResponse, error)
}

type ContestConfigurationOptionsResponse struct {
	Languages              []Language
	Activities             []Activity
	CanCreateOfficialRound bool
}

type ContestConfigurationOptions struct {
	repo ContestConfigurationOptionsRepository
}

func NewContestConfigurationOptions(repo ContestConfigurationOptionsRepository) *ContestConfigurationOptions {
	return &ContestConfigurationOptions{repo: repo}
}

func (s *ContestConfigurationOptions) Execute(ctx context.Context) (*ContestConfigurationOptionsResponse, error) {
	res, err := s.repo.FetchContestConfigurationOptions(ctx)
	if err != nil {
		return nil, err
	}

	res.CanCreateOfficialRound = isAdmin(ctx)

	return res, nil
}
