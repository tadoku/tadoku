package domain

import (
	"time"

	"github.com/google/uuid"
)

// Announcement is a site-wide notification managed by this service.
type Announcement struct {
	ID        uuid.UUID
	Namespace string
	Title     string
	Content   string
	Style     string
	Href      *string
	StartsAt  time.Time
	EndsAt    time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
