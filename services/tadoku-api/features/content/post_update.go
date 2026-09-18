package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type UpdatePostParameters struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (p UpdatePostParameters) Validate() error {
	return validatePostFields(p.Namespace, p.Slug, p.Title, p.Content)
}

func (s *Service) UpdatePost(ctx context.Context, parameters UpdatePostParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	post, err := s.posts.FindPostByID(ctx, parameters.Namespace, parameters.ID)
	if err != nil {
		return err
	}

	contentChanged := post.Title != parameters.Title || post.Content != parameters.Content
	post.Slug = parameters.Slug
	post.Title = parameters.Title
	post.Content = parameters.Content
	post.PublishedAt = parameters.PublishedAt
	post.UpdatedAt = timex.Now()

	return s.posts.UpdatePost(ctx, post, contentChanged)
}
