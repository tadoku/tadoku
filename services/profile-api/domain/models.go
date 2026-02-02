package domain

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents user-specific profile data
type Profile struct {
	ID          uuid.UUID
	UserID      string
	DisplayName string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
