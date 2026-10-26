package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

var (
	ErrStaleSession = &Error{
		"stale_session", errx.NewUnauthorizedError("log in again to change email or password"),
	}
	ErrInvalidRecoveryToken = &Error{
		"invalid_recovery_token", errx.NewInvalidInputError("invalid or expired recovery token"),
	}
	ErrRegistrationThrottled = &Error{
		"throttled", errx.NewThrottledError("too many registration attempts; try again later"),
	}
	ErrRecoveryThrottled = &Error{
		"throttled", errx.NewThrottledError("too many recovery requests; try again later"),
	}
	ErrWrongPassword   = &Error{"wrong_password", errx.NewUnauthorizedError("invalid email or password")}
	ErrEmailTaken      = &Error{"email_taken", errx.NewConflictError("email is already registered")}
	ErrThrottled       = &Error{"throttled", errx.NewThrottledError("too many login attempts; try again later")}
	errAccountNotFound = errors.New("identity account not found")
)

type Account struct {
	ID          uuid.UUID
	Email       Email
	DisplayName DisplayName
	CreatedAt   time.Time
}

type Session interface{ session() }
type GuestSession struct{}

func (GuestSession) session() {}

type AuthenticatedSession struct {
	Account         Account
	AuthenticatedAt time.Time
	ExpiresAt       time.Time
}

func (AuthenticatedSession) session() {}

// RecentLogin reports whether email and password changes are allowed; when it is false they return ErrStaleSession.
func (s AuthenticatedSession) RecentLogin() bool {
	age := timex.Now().Sub(s.AuthenticatedAt)
	return age >= 0 && age <= time.Hour
}

type Cookie struct {
	Name, Value, Domain string
	Expires             time.Time
	MaxAge              int
}
type LoginResult struct {
	Session AuthenticatedSession
	Cookie  Cookie
}
type Registration struct{ Email, DisplayName, Password, IP string }
type Login struct{ Email, Password, IP string }
type Recovery struct{ Email, ReturnTo, IP string }
type RecoveryResult struct{ Emails []mail.Message }
type RecoveryLogin struct {
	Login      LoginResult
	RedirectTo string
}
type ProfileSettings struct{ Email, DisplayName *string }
type ProfileResult struct {
	Account Account
	Emails  []mail.Message
}
type ListAccounts struct {
	Query  string
	Limit  int
	Cursor string
}
type AccountPage struct {
	Accounts   []Account
	NextCursor string
}

type CookieHeader string
type RecoveryToken string

type Reader interface {
	FindAccount(ctx context.Context, userID uuid.UUID) (Account, error)
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	ListAccounts(ctx context.Context, query ListAccounts) (AccountPage, error)
}
