package domain

import (
	"context"
	"fmt"
)

type LanguageUpdateRepository interface {
	UpdateLanguage(ctx context.Context, code string, name string) error
	LanguageExists(ctx context.Context, code string) (bool, error)
}

type LanguageUpdateRequest struct {
	Code string
	Name string
}

type LanguageUpdate struct {
	repo LanguageUpdateRepository
}

func NewLanguageUpdate(repo LanguageUpdateRepository) *LanguageUpdate {
	return &LanguageUpdate{repo: repo}
}

func (s *LanguageUpdate) Execute(ctx context.Context, req *LanguageUpdateRequest) error {
	if isGuest(ctx) {
		return ErrUnauthorized
	}

	if err := requireAdmin(ctx); err != nil {
		return err
	}

	if req.Name == "" || len(req.Name) > 100 {
		return fmt.Errorf("%w: name must be between 1 and 100 characters", ErrRequestInvalid)
	}

	exists, err := s.repo.LanguageExists(ctx, req.Code)
	if err != nil {
		return fmt.Errorf("could not check if language exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	if err := s.repo.UpdateLanguage(ctx, req.Code, req.Name); err != nil {
		return fmt.Errorf("could not update language: %w", err)
	}

	return nil
}
