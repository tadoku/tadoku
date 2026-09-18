package content

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var (
	ErrInvalidPost       = errx.NewInvalidInputError("invalid post")
	ErrPostAlreadyExists = errx.NewConflictError("post already exists")
)

type CreatePostParameters struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

func (p CreatePostParameters) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidPost
	}
	return validatePostFields(p.Namespace, p.Slug, p.Title, p.Content)
}

func validatePostFields(namespace, slug, title, content string) error {
	if namespace == "" || title == "" || content == "" ||
		utf8.RuneCountInString(slug) <= 1 || slug != strings.ToLower(slug) {
		return ErrInvalidPost
	}
	return nil
}

func (s *Service) CreatePost(ctx context.Context, parameters CreatePostParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	now := timex.Now()
	item := &Post{
		ID:          parameters.ID,
		Namespace:   parameters.Namespace,
		Slug:        parameters.Slug,
		Title:       parameters.Title,
		Content:     parameters.Content,
		PublishedAt: parameters.PublishedAt,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	return s.posts.CreatePost(ctx, item)
}
