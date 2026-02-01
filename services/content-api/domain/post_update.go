package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PostUpdateRepository defines the repository interface for updating posts.
type PostUpdateRepository interface {
	GetPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
}

// PostUpdateRequest contains the input data for updating a post.
type PostUpdateRequest struct {
	Namespace   string `validate:"required"`
	Slug        string `validate:"required,gt=1,lowercase"`
	Title       string `validate:"required"`
	Content     string `validate:"required"`
	PublishedAt *time.Time
}

// PostUpdateResponse contains the result of updating a post.
type PostUpdateResponse struct {
	Post *Post
}

// PostUpdate is the service for updating posts.
type PostUpdate struct {
	repo     PostUpdateRepository
	validate *validator.Validate
}

// NewPostUpdate creates a new PostUpdate service.
func NewPostUpdate(repo PostUpdateRepository) *PostUpdate {
	return &PostUpdate{
		repo:     repo,
		validate: validator.New(),
	}
}

// Execute updates an existing post.
// It validates the request, checks authorization, and persists the changes.
func (s *PostUpdate) Execute(ctx context.Context, id uuid.UUID, req *PostUpdateRequest) (*PostUpdateResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPost, err)
	}

	post, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	post.Namespace = req.Namespace
	post.Slug = req.Slug
	post.Title = req.Title
	post.Content = req.Content
	post.PublishedAt = req.PublishedAt
	post.UpdatedAt = time.Now()

	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}

	return &PostUpdateResponse{Post: post}, nil
}
