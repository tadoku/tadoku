package pagecreate

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type PageRepository interface {
	CreatePage(context.Context, *PageCreateRequest) (*PageCreateResponse, error)
}

type Service interface {
	CreatePage(context.Context, *PageCreateRequest) (*PageCreateResponse, error)
}

type service struct {
	pr       PageRepository
	validate *validator.Validate
}

func NewService(pr PageRepository) Service {
	return &service{
		pr:       pr,
		validate: validator.New(),
	}
}

func (s *service) CreatePage(ctx context.Context, req *PageCreateRequest) (*PageCreateResponse, error) {
	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPage, err)
	}

	return nil, nil
}
