package domain

import (
	"context"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
	if commondomain.IsRole(ctx, commondomain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	languages, err := s.repo.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list languages: %w", err)
	}

	return languages, nil
}
