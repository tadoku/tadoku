package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PostCreateRepository interface {
	CreatePost(ctx context.Context, post *Post) error
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
	Post *Post
}

type PostCreate struct {
	repo     PostCreateRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewPostCreate(repo PostCreateRepository, clock commondomain.Clock) *PostCreate {
	return &PostCreate{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

func (s *PostCreate) Execute(ctx context.Context, req *PostCreateRequest) (*PostCreateResponse, error) {
	if !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrForbidden
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPost, err)
	}

	now := s.clock.Now()
	post := &Post{
		ID:          req.ID,
		Namespace:   req.Namespace,
		Slug:        req.Slug,
		Title:       req.Title,
		Content:     req.Content,
		PublishedAt: req.PublishedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	return &PostCreateResponse{Post: post}, nil
}
