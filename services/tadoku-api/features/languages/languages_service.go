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
