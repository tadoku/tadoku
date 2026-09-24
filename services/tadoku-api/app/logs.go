package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type LogConfigurationOptions struct {
	Languages            []languages.Language
	Activities           []activities.Activity
	Units                []logs.Unit
	UserLanguageCodes    []string
	ScoringEngineEnabled bool
}

func (a *Application) LogConfigurationOptions(ctx context.Context) (*LogConfigurationOptions, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	userID, err := identity.RequireCallerID(ctx)
	if err != nil {
		return nil, err
	}

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}
	options, err := a.logs.ConfigurationOptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &LogConfigurationOptions{
		Languages:            languages,
		Activities:           activities.All(),
		Units:                options.Units,
		UserLanguageCodes:    options.UserLanguageCodes,
		ScoringEngineEnabled: options.ScoringEngineEnabled,
	}, nil
}

func (a *Application) LogTagSuggestions(ctx context.Context, query string) ([]logs.TagSuggestion, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	userID, err := identity.RequireCallerID(ctx)
	if err != nil {
		return nil, err
	}

	return a.logs.TagSuggestions(ctx, userID, query)
}
