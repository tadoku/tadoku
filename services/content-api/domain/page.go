package domain

import (
	"time"

	"github.com/google/uuid"
)

// Page is a web page whose content is managed by this service.
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
