package postquery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var ErrPostNotFound = errors.New("post not found")
var ErrRequestInvalid = errors.New("request is invalid")

type PostRepository interface {
	FindBySlug(context.Context, *PostFindRequest) (*PostFindResponse, error)
	ListPosts(context.Context, *PostListRequest) (*PostListResponse, error)
}

type Service interface {
	FindBySlug(context.Context, *PostFindRequest) (*PostFindResponse, error)
	ListPosts(context.Context, *PostListRequest) (*PostListResponse, error)
}

type service struct {
	pr       PostRepository
	validate *validator.Validate
}

func NewService(pr PostRepository) Service {
	return &service{
		pr:       pr,
		validate: validator.New(),
	}
}

type PostFindRequest struct {
	Slug      string `validate:"required"`
	Namespace string `validate:"required"`
}

type PostFindResponse struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (s *service) FindBySlug(ctx context.Context, req *PostFindRequest) (*PostFindResponse, error) {
	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
	}

	post, err := s.pr.FindBySlug(ctx, req)
	if err != nil {
		return nil, err
	}

	// TODO: Extract out time.Now() into a clock provider so it can be mocked
	if post.PublishedAt == nil || post.PublishedAt.After(time.Now()) {
		return nil, fmt.Errorf("post is not published yet: %w", ErrPostNotFound)
	}

	return post, nil
}

type PostListRequest struct {
	Namespace     string `validate:"required"`
	IncludeDrafts bool
	PageSize      int
	Page          int
}

type PostListResponse struct {
	Posts         []PostListEntry
	TotalSize     int
	NextPageToken string
}

type PostListEntry struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *service) ListPosts(ctx context.Context, req *PostListRequest) (*PostListResponse, error) {
	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.pr.ListPosts(ctx, req)
}
