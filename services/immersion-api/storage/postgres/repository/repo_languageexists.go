package repository

import (
	"context"
	"fmt"
)

func (r *Repository) LanguageExists(ctx context.Context, code string) (bool, error) {
	languages, err := r.q.GetLanguagesByCode(ctx, []string{code})
	if err != nil {
		return false, fmt.Errorf("could not check if language exists: %w", err)
	}
	return len(languages) > 0, nil
}
