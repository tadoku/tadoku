package domain

import (
	"context"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
	if commondomain.IsRole(ctx, commondomain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	return s.repo.FetchLogConfigurationOptions(ctx)
}
