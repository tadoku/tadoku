package postquery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrPostNotFound = errors.New("post not found")

type PostRepository interface {
	FindBySlug(context.Context, string, string) (*PostFindResponse, error)
	// ListPosts(context.Context) (*PostListResponse, error)
}

type Service interface {
	FindBySlug(context.Context, string, string) (*PostFindResponse, error)
	// ListPosts(context.Context) (*PostListResponse, error)
}

type service struct {
	pr PostRepository
}

func NewService(pr PostRepository) Service {
	return &service{
		pr: pr,
	}
}

type PostFindResponse struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (s *service) FindBySlug(ctx context.Context, namespace, slug string) (*PostFindResponse, error) {
	post, err := s.pr.FindBySlug(ctx, namespace, slug)
	if err != nil {
		return nil, err
	}

	// TODO: Extract out time.Now() into a clock provider so it can be mocked
	if post.PublishedAt == nil || post.PublishedAt.After(time.Now()) {
		return nil, fmt.Errorf("post is not published yet: %w", ErrPostNotFound)
	}

	return post, nil
}

// type PostListResponse struct {
// 	Posts []PostListEntry
// }

// type PostListEntry struct {
// 	ID          uuid.UUID
// 	Slug        string
// 	Title       string
// 	PublishedAt *time.Time
// 	CreatedAt   time.Time
// 	UpdatedAt   time.Time
// }

// func (s *service) ListPosts(ctx context.Context) (*PostListResponse, error) {
// 	return s.pr.ListPosts(ctx)
// }
