package domain

import (
	"context"
	"fmt"
)

type LanguageListRepository interface {
	ListLanguages(ctx context.Context) ([]Language, error)
}

type LanguageList struct {
	repo LanguageListRepository
}

func NewLanguageList(repo LanguageListRepository) *LanguageList {
	return &LanguageList{repo: repo}
}

func (s *LanguageList) Execute(ctx context.Context) ([]Language, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	languages, err := s.repo.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list languages: %w", err)
	}

	return languages, nil
}
