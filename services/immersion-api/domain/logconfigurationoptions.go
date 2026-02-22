package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogConfigurationOptionsRepository interface {
	FetchLogConfigurationOptions(ctx context.Context, userID uuid.UUID) (*LogConfigurationOptionsResponse, error)
}

type LogConfigurationOptionsResponse struct {
	Languages         []Language
	Activities        []Activity
	Units             []Unit
	UserLanguageCodes []string
}

type LogConfigurationOptions struct {
	repo LogConfigurationOptionsRepository
}

func NewLogConfigurationOptions(repo LogConfigurationOptionsRepository) *LogConfigurationOptions {
	return &LogConfigurationOptions{repo: repo}
}

func (s *LogConfigurationOptions) Execute(ctx context.Context) (*LogConfigurationOptionsResponse, error) {
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil || session.Subject == "" {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(session.Subject)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return s.repo.FetchLogConfigurationOptions(ctx, userID)
}
