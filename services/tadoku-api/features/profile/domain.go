package profile

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrLocalUserNotFound         = errx.NewNotFoundError("local user not found")
	ErrIdentityNotFound          = errx.NewNotFoundError("identity not found")
	ErrAccountDeletionInProgress = errx.NewConflictError("account deletion in progress")
)

type UserDeletionState struct {
	DeletionLocked bool
	Deleted        bool
}

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

type PublicProfile struct {
	DisplayName string
	CreatedAt   time.Time
}
