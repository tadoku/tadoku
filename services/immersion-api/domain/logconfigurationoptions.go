package domain

import (
	"context"
)

type LogConfigurationOptionsRepository interface {
	FetchLogConfigurationOptions(ctx context.Context) (*LogConfigurationOptionsResponse, error)
}

type LogConfigurationOptionsResponse struct {
	Languages  []Language
	Activities []Activity
	Units      []Unit
	Tags       []Tag
}

type LogConfigurationOptions struct {
	repo LogConfigurationOptionsRepository
}

func NewLogConfigurationOptions(repo LogConfigurationOptionsRepository) *LogConfigurationOptions {
	return &LogConfigurationOptions{repo: repo}
}

func (s *LogConfigurationOptions) Execute(ctx context.Context) (*LogConfigurationOptionsResponse, error) {
	if isGuest(ctx) {
		return nil, ErrUnauthorized
	}

	return s.repo.FetchLogConfigurationOptions(ctx)
}
