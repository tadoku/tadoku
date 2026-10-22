package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
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
	ErrNotImplemented  = errx.NewNotImplementedError()
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

type Accounts interface {
	Reader
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	Register(ctx context.Context, registration Registration) (LoginResult, error)
	Login(ctx context.Context, login Login) (LoginResult, error)
	Logout(ctx context.Context, cookieHeader CookieHeader) (Cookie, error)
	Session(ctx context.Context, cookieHeader CookieHeader) (Session, error)
	RequestRecovery(ctx context.Context, recovery Recovery) (RecoveryResult, error)
	ConsumeRecovery(ctx context.Context, recoveryToken RecoveryToken) (RecoveryLogin, error)
	UpdateProfile(ctx context.Context, cookieHeader CookieHeader, settings ProfileSettings) (ProfileResult, error)
	UpdatePassword(ctx context.Context, cookieHeader CookieHeader, password string) error
}
