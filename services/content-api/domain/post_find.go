package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PostFindRepository defines the repository interface for PostFind.
type PostFindRepository interface {
	FindPostBySlug(ctx context.Context, namespace, slug string) (*Post, error)
}

type PostFindRequest struct {
	Namespace string `validate:"required"`
	Slug      string `validate:"required"`
}

type PostFindResponse struct {
	Post *Post
}

type PostFind struct {
	repo     PostFindRepository
	validate *validator.Validate
	clock    commondomain.Clock
}

func NewPostFind(repo PostFindRepository, clock commondomain.Clock) *PostFind {
	return &PostFind{
		repo:     repo,
		validate: validator.New(),
		clock:    clock,
	}
}

func (s *PostFind) Execute(ctx context.Context, req *PostFindRequest) (*PostFindResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	post, err := s.repo.FindPostBySlug(ctx, req.Namespace, req.Slug)
	if err != nil {
		return nil, err
	}

	if post.PublishedAt == nil || post.PublishedAt.After(s.clock.Now()) {
		return nil, fmt.Errorf("post is not published yet: %w", ErrPostNotFound)
	}

	return &PostFindResponse{Post: post}, nil
}
