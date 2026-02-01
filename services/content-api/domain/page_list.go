package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageListRepository defines the repository interface for listing pages.
type PageListRepository interface {
	ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PageListResult, error)
}

// PageListResult contains the paginated list of pages from the repository.
type PageListResult struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

// PageListRequest contains the input data for listing pages.
type PageListRequest struct {
	Namespace     string `validate:"required"`
	IncludeDrafts bool
	PageSize      int
	Page          int
}

// PageListResponse contains the result of listing pages.
type PageListResponse struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

// PageList is the service for listing pages.
type PageList struct {
	repo     PageListRepository
	validate *validator.Validate
}

// NewPageList creates a new PageList service.
func NewPageList(repo PageListRepository) *PageList {
	return &PageList{
		repo:     repo,
		validate: validator.New(),
	}
}

// Execute lists pages for a namespace.
// Only admins can access this endpoint.
func (s *PageList) Execute(ctx context.Context, req *PageListRequest) (*PageListResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	// Apply defaults
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := s.repo.ListPages(ctx, req.Namespace, req.IncludeDrafts, pageSize, req.Page)
	if err != nil {
		return nil, err
	}

	return &PageListResponse{
		Pages:         result.Pages,
		TotalSize:     result.TotalSize,
		NextPageToken: result.NextPageToken,
	}, nil
}
