package languages

import "context"

type Service struct {
	languages *LanguagesRepository
}

func NewService(languages *LanguagesRepository) *Service {
	return &Service{languages: languages}
}

func (s *Service) ListLanguages(ctx context.Context) ([]Language, error) {
	return s.languages.ListLanguages(ctx)
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
