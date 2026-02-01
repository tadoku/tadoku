package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageCreateRepository defines the repository interface for creating pages.
// This is a minimal interface containing only what PageCreate needs.
type PageCreateRepository interface {
	CreatePage(ctx context.Context, page *Page) error
}

// PageCreateRequest contains the input data for creating a page.
type PageCreateRequest struct {
	ID          uuid.UUID `validate:"required"`
	Namespace   string    `validate:"required"`
	Slug        string    `validate:"required,gt=1,lowercase"`
	Title       string    `validate:"required"`
	HTML        string    `validate:"required"`
	PublishedAt *time.Time
}

// PageCreateResponse contains the result of creating a page.
type PageCreateResponse struct {
	Page *Page
}

// PageCreate is the service for creating pages.
type PageCreate struct {
	repo     PageCreateRepository
	validate *validator.Validate
}

// NewPageCreate creates a new PageCreate service.
func NewPageCreate(repo PageCreateRepository) *PageCreate {
	return &PageCreate{
		repo:     repo,
		validate: validator.New(),
	}
}

// Execute creates a new page.
// It validates the request, checks authorization, and persists the page.
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
