package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) LanguageExists(ctx context.Context, code string) (bool, error) {
	languages, err := r.q.GetLanguagesByCode(ctx, []string{code})
	if err != nil {
		return false, fmt.Errorf("could not check if language exists: %w", err)
	}
	return len(languages) > 0, nil
}

func (r *Repository) LanguagesExist(ctx context.Context, codes []string) (bool, error) {
	if len(codes) == 0 {
		return true, nil
	}

	languages, err := r.q.GetLanguagesByCode(ctx, codes)
	if err != nil {
		return false, fmt.Errorf("could not check if languages exist: %w", err)
	}

	return allRequestedLanguagesExist(codes, languages), nil
}

func allRequestedLanguagesExist(codes []string, languages []postgres.Language) bool {
	missingCodes := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		missingCodes[code] = struct{}{}
	}

	for _, language := range languages {
		delete(missingCodes, language.Code)
	}

	return len(missingCodes) == 0
}
