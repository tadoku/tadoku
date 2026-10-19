package tokens

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

var ErrNotImplemented = errx.NewNotImplementedError()

type Scope string

const EmailRead Scope = "email:read"

type Token struct {
	ID                    uuid.UUID
	Name                  string
	Scopes                []Scope
	CreatedAt, ExpiresAt  time.Time
	LastUsedAt, RevokedAt *time.Time
}
type Principal struct {
	Tenant  tenant.Key
	TokenID uuid.UUID
	Scopes  []Scope
}
type Create struct {
	Name      string
	Scopes    []Scope
	ExpiresIn time.Duration
}
type Issued struct {
	Token  Token
	Secret string
}

type Verifier interface {
	Verify(context.Context, string, tenant.Deployment) (Principal, error)
}
type Manager interface {
	Create(context.Context, Create) (Issued, error)
	List(context.Context) ([]Token, error)
	Revoke(context.Context, uuid.UUID) error
}
