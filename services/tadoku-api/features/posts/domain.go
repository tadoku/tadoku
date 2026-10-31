package posts

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Post struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// IsPublishedAt is the publication rule. ListPosts applies it in SQL to the cutoff it receives, so change both together.
func (p Post) IsPublishedAt(t time.Time) bool {
	return p.PublishedAt != nil && !p.PublishedAt.After(t)
}

var (
	ErrPostNotFound      = errx.NewNotFoundError("post not found")
	ErrPostAlreadyExists = errx.NewConflictError("post already exists")
)

type PostContent struct {
	ID        uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
}

type PostVersion struct {
	PostContent
	Version int
}

type postParameters struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
}

type CreatePostParameters = postParameters
type UpdatePostParameters = postParameters

func (p postParameters) Validate() error {
	if p.ID == uuid.Nil {
		return errx.NewInvalidInputError("id is required")
	}
	return validatePostFields(p.Namespace, p.Slug, p.Title, p.Content)
}

func validatePostFields(namespace, slug, title, content string) error {
	if namespace == "" {
		return errx.NewInvalidInputError("namespace is required")
	}
	if utf8.RuneCountInString(slug) <= 1 {
		return errx.NewInvalidInputError("slug must be at least 2 characters")
	}
	if slug != strings.ToLower(slug) {
		return errx.NewInvalidInputError("slug must be lowercase")
	}
	if title == "" {
		return errx.NewInvalidInputError("title is required")
	}
	if content == "" {
		return errx.NewInvalidInputError("content is required")
	}
	return nil
}

type PostList struct {
	Posts         []Post
	TotalSize     int
	NextPageToken string
}
