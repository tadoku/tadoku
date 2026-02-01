package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PageListRepository defines the repository interface for PageList.
type PageListRepository interface {
	ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PageListResult, error)
}

type PageListResult struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

type PageListRequest struct {
	Namespace     string `validate:"required"`
	IncludeDrafts bool
	PageSize      int
	Page          int
}

type PageListResponse struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

type PageList struct {
	repo     PageListRepository
	validate *validator.Validate
}

func NewPageList(repo PageListRepository) *PageList {
	return &PageList{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *PageList) Execute(ctx context.Context, req *PageListRequest) (*PageListResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

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
