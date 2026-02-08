package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type PostListRepository interface {
	ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PostListResult, error)
}

type PostListResult struct {
	Posts         []Post
	TotalSize     int
	NextPageToken string
}

type PostListRequest struct {
	Namespace     string `validate:"required"`
	IncludeDrafts bool
	PageSize      int
	Page          int
}

type PostListResponse struct {
	Posts         []Post
	TotalSize     int
	NextPageToken string
}

type PostList struct {
	repo     PostListRepository
	validate *validator.Validate
}

func NewPostList(repo PostListRepository) *PostList {
	return &PostList{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *PostList) Execute(ctx context.Context, req *PostListRequest) (*PostListResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	if req.IncludeDrafts && !isAdmin(ctx) {
		return nil, ErrForbidden
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := s.repo.ListPosts(ctx, req.Namespace, req.IncludeDrafts, pageSize, req.Page)
	if err != nil {
		return nil, err
	}

	return &PostListResponse{
		Posts:         result.Posts,
		TotalSize:     result.TotalSize,
		NextPageToken: result.NextPageToken,
	}, nil
}
