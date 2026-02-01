package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageCreateRepository defines the repository interface for PageCreate.
type PageCreateRepository interface {
	CreatePage(ctx context.Context, page *Page) error
}

type PageCreateRequest struct {
	ID          uuid.UUID `validate:"required"`
	Namespace   string    `validate:"required"`
	Slug        string    `validate:"required,gt=1,lowercase"`
	Title       string    `validate:"required"`
	HTML        string    `validate:"required"`
	PublishedAt *time.Time
}

type PageCreateResponse struct {
	Page *Page
}

type PageCreate struct {
	repo     PageCreateRepository
	validate *validator.Validate
}

func NewPageCreate(repo PageCreateRepository) *PageCreate {
	return &PageCreate{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *PageCreate) Execute(ctx context.Context, req *PageCreateRequest) (*PageCreateResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPage, err)
	}

	now := time.Now()
	page := &Page{
		ID:          req.ID,
		Namespace:   req.Namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		HTML:        req.HTML,
		PublishedAt: req.PublishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreatePage(ctx, page); err != nil {
		return nil, err
	}

	return &PageCreateResponse{Page: page}, nil
}
