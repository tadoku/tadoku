package domain

import (
	"time"

	"github.com/google/uuid"
)

// Page represents a content page in the domain layer.
// This is separate from the database model to decouple business logic from storage.
type Page struct {
	ID          uuid.UUID
	Namespace   string
	Slug        string
	Title       string
	HTML        string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
