package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageFindRepository defines the repository interface for PageFind.
type PageFindRepository interface {
	FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error)
}

type PageFindRequest struct {
	Namespace string `validate:"required"`
	Slug      string `validate:"required"`
}

type PageFindResponse struct {
	Page *Page
}

type PageFind struct {
	repo     PageFindRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewPageFind(repo PageFindRepository, clock commondomain.Clock) *PageFind {
	return &PageFind{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

func (s *PageFind) Execute(ctx context.Context, req *PageFindRequest) (*PageFindResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	page, err := s.repo.FindPageBySlug(ctx, req.Namespace, req.Slug)
	if err != nil {
		return nil, err
	}

	if page.PublishedAt == nil || page.PublishedAt.After(s.clock.Now()) {
		return nil, fmt.Errorf("page is not published yet: %w", ErrPageNotFound)
	}

	return &PageFindResponse{Page: page}, nil
}
