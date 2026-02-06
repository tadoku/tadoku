package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) UpdateLanguage(ctx context.Context, code string, name string) error {
	err := r.q.UpdateLanguage(ctx, postgres.UpdateLanguageParams{
		Code: code,
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("could not update language: %w", err)
	}
	return nil
}
