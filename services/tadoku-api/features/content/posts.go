package content

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

var ErrPostNotFound = errx.NewNotFoundError("post not found")
