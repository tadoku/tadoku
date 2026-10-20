package profile

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrLocalUserNotFound         = errx.NewNotFoundError("local user not found")
	ErrAccountDeletionInProgress = errx.NewConflictError("account deletion in progress")
)

type userDeletionState struct {
	DeletionLocked bool
	Deleted        bool
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
