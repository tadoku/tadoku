package domain

import (
	"time"

	"github.com/google/uuid"
)

// Post is a blog post managed by this service.
type Post struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
