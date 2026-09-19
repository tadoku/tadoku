package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
)

func (a *Application) ListLanguages(ctx context.Context) ([]languages.Language, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.languages.ListLanguages(ctx)
}
