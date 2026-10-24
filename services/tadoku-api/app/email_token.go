package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/authz/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

var ErrEmailTenantNotServed = tokens.ErrTenantNotServed

type emailReaderKey struct{}

func (a *Application) AuthenticateEmailReader(ctx context.Context, raw string) (context.Context, error) {
	if a.tokens == nil {
		return nil, errx.NewUnauthorizedError("access token authentication unavailable")
	}

	principal, err := a.tokens.Verify(ctx, raw, a.deployment)
	if err != nil {
		return nil, err
	}

	ctx = tenant.WithKey(ctx, principal.Tenant)
	return context.WithValue(ctx, emailReaderKey{}, principal), nil
}

func (a *Application) hasEmailReader(ctx context.Context) bool {
	principal, ok := ctx.Value(emailReaderKey{}).(tokens.Principal)
	key, present := tenant.FromContext(ctx)
	return ok && present && principal.Tenant == key &&
		len(principal.Scopes) == 1 && principal.Scopes[0] == tokens.EmailRead
}

func (a *Application) requireEmailReader(ctx context.Context) error {
	if a.hasEmailReader(ctx) {
		return nil
	}
	return a.roles.RequireAdmin(ctx)
}
