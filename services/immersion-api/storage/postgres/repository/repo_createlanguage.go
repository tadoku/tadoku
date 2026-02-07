package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) CreateLanguage(ctx context.Context, code string, name string) error {
	err := r.q.CreateLanguage(ctx, postgres.CreateLanguageParams{
		Code: code,
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("could not create language: %w", err)
	}
	return nil
}
