// Package posts owns posts and their persistence.
package posts

import (
	"time"

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

var (
	ErrInvalidNamespace  = errx.NewInvalidInputError("namespace is required")
	ErrInvalidPagination = errx.NewInvalidInputError("invalid pagination")
	ErrPostNotFound      = errx.NewNotFoundError("post not found")
)

type Service struct {
	posts *PostsRepository
}

func NewService(posts *PostsRepository) *Service {
	return &Service{posts: posts}
}

type PostVersion struct {
	ID        uuid.UUID
	Version   int
	Title     string
	Content   string
	CreatedAt time.Time
}
