package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PostUpdateRepository interface {
	GetPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	UpdatePostMetadata(ctx context.Context, post *Post) error
}

type PostUpdateRequest struct {
	Namespace   string `validate:"required"`
	Slug        string `validate:"required,gt=1,lowercase"`
	Title       string `validate:"required"`
	Content     string `validate:"required"`
	PublishedAt *time.Time
}

type PostUpdateResponse struct {
	Post *Post
}

type PostUpdate struct {
	repo     PostUpdateRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewPostUpdate(repo PostUpdateRepository, clock commondomain.Clock) *PostUpdate {
	return &PostUpdate{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

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

	contentChanged := post.Title != req.Title || post.Content != req.Content

	post.Namespace = req.Namespace
	post.Slug = req.Slug
	post.Title = req.Title
	post.Content = req.Content
	post.PublishedAt = req.PublishedAt
	post.UpdatedAt = s.clock.Now()

	if contentChanged {
		err = s.repo.UpdatePost(ctx, post)
	} else {
		err = s.repo.UpdatePostMetadata(ctx, post)
	}
	if err != nil {
		return nil, err
	}

	return &PostUpdateResponse{Post: post}, nil
}
