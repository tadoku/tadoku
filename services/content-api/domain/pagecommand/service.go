package pagecommand

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var ErrPageAlreadyExists = errors.New("page with given id already exists")
var ErrInvalidPage = errors.New("unable to validate page")

type PageCreateRequest struct {
	ID          uuid.UUID `validate:"required"`
	Slug        string    `validate:"required,gt=1,lowercase"`
	Title       string    `validate:"required"`
	Html        string    `validate:"required"`
	PublishedAt *time.Time
}

type PageCreateResponse struct {
	ID    uuid.UUID
	Slug  string
	Title string
	Html  string
}

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

	return s.pr.CreatePage(ctx, req)
}
