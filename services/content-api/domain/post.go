package domain

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a blog post in the domain layer.
// This is separate from the database model to decouple business logic from storage.
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
