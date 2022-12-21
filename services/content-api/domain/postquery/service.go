package postquery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrPostNotFound = errors.New("post not found")
var ErrRequestInvalid = errors.New("request is invalid")
var ErrForbidden = errors.New("not allowed")

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
	clock    domain.Clock
}

func NewService(pr PostRepository, clock domain.Clock) Service {
	return &service{
		pr:       pr,
		validate: validator.New(),
		clock:    clock,
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

	if post.PublishedAt == nil || post.PublishedAt.After(s.clock.Now()) {
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

	if req.IncludeDrafts && !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.pr.ListPosts(ctx, req)
}
