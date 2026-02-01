package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageFindRepository defines the repository interface for finding pages.
type PageFindRepository interface {
	FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error)
}

// PageFindRequest contains the input data for finding a page.
type PageFindRequest struct {
	Namespace string `validate:"required"`
	Slug      string `validate:"required"`
}

// PageFindResponse contains the result of finding a page.
type PageFindResponse struct {
	Page *Page
}

// PageFind is the service for finding pages by slug.
type PageFind struct {
	repo     PageFindRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

// NewPageFind creates a new PageFind service.
func NewPageFind(repo PageFindRepository, clock commondomain.Clock) *PageFind {
	return &PageFind{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

// Execute finds a page by namespace and slug.
// Returns ErrPageNotFound if the page doesn't exist or is not yet published.
func (s *PageFind) Execute(ctx context.Context, req *PageFindRequest) (*PageFindResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	page, err := s.repo.FindPageBySlug(ctx, req.Namespace, req.Slug)
	if err != nil {
		return nil, err
	}

	// Check if page is published
	if page.PublishedAt == nil || page.PublishedAt.After(s.clock.Now()) {
		return nil, fmt.Errorf("page is not published yet: %w", ErrPageNotFound)
	}

	return &PageFindResponse{Page: page}, nil
}
