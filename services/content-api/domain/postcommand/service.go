package postcommand

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrPostAlreadyExists = errors.New("post with given id already exists")
var ErrPostNotFound = errors.New("post not found")
var ErrInvalidPost = errors.New("unable to validate page")
var ErrForbidden = errors.New("not allowed")

type PostRepository interface {
	CreatePost(context.Context, *PostCreateRequest) (*PostCreateResponse, error)
	UpdatePost(context.Context, uuid.UUID, *PostUpdateRequest) (*PostUpdateResponse, error)
}

type Service interface {
	CreatePost(context.Context, *PostCreateRequest) (*PostCreateResponse, error)
	UpdatePost(context.Context, uuid.UUID, *PostUpdateRequest) (*PostUpdateResponse, error)
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

type PostCreateRequest struct {
	ID          uuid.UUID `validate:"required"`
	Namespace   string    `validate:"required"`
	Slug        string    `validate:"required,gt=1,lowercase"`
	Title       string    `validate:"required"`
	Content     string    `validate:"required"`
	PublishedAt *time.Time
}

type PostCreateResponse struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (s *service) CreatePost(ctx context.Context, req *PostCreateRequest) (*PostCreateResponse, error) {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrForbidden
	}

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPost, err)
	}

	return s.pr.CreatePost(ctx, req)
}

type PostUpdateRequest struct {
	Slug        string `validate:"required,gt=1,lowercase"`
	Namespace   string `validate:"required"`
	Title       string `validate:"required"`
	Content     string `validate:"required"`
	PublishedAt *time.Time
}

type PostUpdateResponse struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (s *service) UpdatePost(ctx context.Context, id uuid.UUID, req *PostUpdateRequest) (*PostUpdateResponse, error) {
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrForbidden
	}

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPost, err)
	}

	return s.pr.UpdatePost(ctx, id, req)
}
