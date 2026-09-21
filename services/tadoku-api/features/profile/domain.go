// Package profile owns user profile operations.
package profile

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrLocalUserNotFound         = errx.NewNotFoundError("local user not found")
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

type ActivityScore struct {
	Date    time.Time
	Score   float32
	Updates int
}

type YearlyActivity struct {
	Scores       []ActivityScore
	TotalUpdates int
}

type Score struct {
	LanguageCode string
	LanguageName string
	Score        float32
}

type YearlyScores struct {
	Scores       []Score
	OverallScore float32
}

type ActivitySplitScore struct {
	ActivityID   int
	ActivityName string
	Score        float32
}
