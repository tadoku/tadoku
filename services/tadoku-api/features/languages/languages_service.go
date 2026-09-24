package languages

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Service struct {
	languages *LanguagesRepository
}

func NewService(languages *LanguagesRepository) *Service {
	return &Service{languages: languages}
}

func (s *Service) ListLanguages(ctx context.Context) ([]Language, error) {
	return s.languages.ListLanguages(ctx)
}

func (s *Service) RequireExistingLanguages(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	exist, err := s.languages.LanguagesExist(ctx, codes)
	if err != nil {
		return err
	}
	if !exist {
		return errx.NewInvalidInputError("invalid language codes: one or more languages do not exist")
	}
	return nil
}

func (s *Service) CreateLanguage(ctx context.Context, parameters CreateLanguageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}
	return s.languages.CreateLanguage(ctx, parameters)
}

func (s *Service) UpdateLanguage(ctx context.Context, parameters UpdateLanguageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}
	return s.languages.UpdateLanguage(ctx, parameters)
}
