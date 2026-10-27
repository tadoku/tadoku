package profile

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrLocalUserNotFound         = errx.NewNotFoundError("local user not found")
	ErrAccountDeletionInProgress = errx.NewConflictError("account deletion in progress")
	ErrDeletionNotRequested      = errx.NewConflictError("account deletion was not requested")
)

type userDeletionState struct {
	DeletionLocked bool
	Deleted        bool
}

func (s userDeletionState) requireDeletable() error {
	switch {
	case s.Deleted:
		return ErrLocalUserNotFound
	case s.DeletionLocked:
		return ErrAccountDeletionInProgress
	}
	return nil
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
