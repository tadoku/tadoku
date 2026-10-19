package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var ErrNotImplemented = errx.NewNotImplementedError()

type AccountApp string

const (
	V1 AccountApp = "v1"
	V2 AccountApp = "v2"
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

type SessionResult struct {
	AccountApp AccountApp
	Session    Session
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
type Registration struct{ Email, DisplayName, Password string }
type Login struct{ Email, Password, IP string }
type Recovery struct{ Email, ReturnTo string }
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

type Reader interface {
	FindAccount(context.Context, uuid.UUID) (Account, error)
	UserExists(context.Context, uuid.UUID) (bool, error)
	ListAccounts(context.Context, ListAccounts) (AccountPage, error)
}

type Accounts interface {
	Reader
	DeleteAccount(context.Context, uuid.UUID) error
	Register(context.Context, Registration) (LoginResult, error)
	Login(context.Context, Login) (LoginResult, error)
	Logout(context.Context, string) (Cookie, error)
	Session(context.Context, string) (SessionResult, error)
	CheckSession(context.Context, string) (Session, error)
	RequestRecovery(context.Context, Recovery) (RecoveryResult, error)
	ConsumeRecovery(context.Context, string) (RecoveryLogin, error)
	UpdateProfile(context.Context, string, ProfileSettings) (ProfileResult, error)
	UpdatePassword(context.Context, string, string) error
}
