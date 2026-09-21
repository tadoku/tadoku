// Package profile owns user profile operations.
package profile

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrInvalidSignedUser         = errors.New("invalid signed user identity")
	ErrLocalUserNotFound         = errx.NewNotFoundError("local user not found")
	ErrAccountDeletionInProgress = errx.NewConflictError("account deletion in progress")
)

type SignedUser struct {
	ID          uuid.UUID
	DisplayName string

	sessionCreatedAt time.Time
}

func (u SignedUser) SessionCreatedAt() time.Time { return u.sessionCreatedAt }

type CachedUser struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

type User struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
	Role        string
}

type UserList struct {
	Users     []User
	TotalSize int
}
