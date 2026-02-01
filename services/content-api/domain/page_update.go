package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageUpdateRepository defines the repository interface for PageUpdate.
type PageUpdateRepository interface {
	GetPageByID(ctx context.Context, id uuid.UUID) (*Page, error)
	UpdatePage(ctx context.Context, page *Page) error
}

type PageUpdateRequest struct {
	Namespace   string `validate:"required"`
	Slug        string `validate:"required,gt=1,lowercase"`
	Title       string `validate:"required"`
	HTML        string `validate:"required"`
	PublishedAt *time.Time
}

type PageUpdateResponse struct {
	Page *Page
}

type PageUpdate struct {
	repo     PageUpdateRepository
	validate *validator.Validate
}

func NewPageUpdate(repo PageUpdateRepository) *PageUpdate {
	return &PageUpdate{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *PageUpdate) Execute(ctx context.Context, id uuid.UUID, req *PageUpdateRequest) (*PageUpdateResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPage, err)
	}

	page, err := s.repo.GetPageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	page.Namespace = req.Namespace
	page.Slug = req.Slug
	page.Title = req.Title
	page.HTML = req.HTML
	page.PublishedAt = req.PublishedAt
	page.UpdatedAt = time.Now()

	if err := s.repo.UpdatePage(ctx, page); err != nil {
		return nil, err
	}

	return &PageUpdateResponse{Page: page}, nil
}
