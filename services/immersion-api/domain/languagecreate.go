package domain

import (
	"context"
	"fmt"
)

type LanguageCreateRepository interface {
	CreateLanguage(ctx context.Context, code string, name string) error
	LanguageExists(ctx context.Context, code string) (bool, error)
}

type LanguageCreateRequest struct {
	Code string
	Name string
}

type LanguageCreate struct {
	repo LanguageCreateRepository
}

func NewLanguageCreate(repo LanguageCreateRepository) *LanguageCreate {
	return &LanguageCreate{repo: repo}
}

func (s *LanguageCreate) Execute(ctx context.Context, req *LanguageCreateRequest) error {
	if isGuest(ctx) {
		return ErrUnauthorized
	}

	if err := requireAdmin(ctx); err != nil {
		return err
	}

	if req.Code == "" || len(req.Code) > 10 {
		return fmt.Errorf("%w: code must be between 1 and 10 characters", ErrRequestInvalid)
	}

	if req.Name == "" || len(req.Name) > 100 {
		return fmt.Errorf("%w: name must be between 1 and 100 characters", ErrRequestInvalid)
	}

	exists, err := s.repo.LanguageExists(ctx, req.Code)
	if err != nil {
		return fmt.Errorf("could not check if language exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: language with code '%s' already exists", ErrConflict, req.Code)
	}

	if err := s.repo.CreateLanguage(ctx, req.Code, req.Name); err != nil {
		return fmt.Errorf("could not create language: %w", err)
	}

	return nil
}
