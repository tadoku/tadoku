package app

import (
	"context"
	"errors"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/callbackauth"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

var ErrIdentityTenantNotServed = errors.New("tenant not served by this deployment")

func (a *Application) IdentityRegister(ctx context.Context, registration identity.Registration) (identity.LoginResult, error) {
	var result identity.LoginResult
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var err error
		result, err = a.identity.Register(ctx, registration)
		return err
	})
	return result, err
}

func (a *Application) IdentityLogin(ctx context.Context, login identity.Login) (identity.LoginResult, error) {
	return a.identity.Login(ctx, login)
}

func (a *Application) IdentityLogout(ctx context.Context, cookieHeader identity.CookieHeader) (identity.Cookie, error) {
	return a.identity.Logout(ctx, cookieHeader)
}

func (a *Application) IdentitySession(ctx context.Context, cookieHeader identity.CookieHeader) (identity.SessionResult, error) {
	return a.identity.Session(ctx, cookieHeader)
}

func (a *Application) IdentityCheckSession(
	ctx context.Context,
	branch string,
	cookieHeader identity.CookieHeader,
) (identity.Session, error) {
	if !callbackauth.IsAuthenticated(ctx) {
		return nil, errx.NewUnauthorizedError("callback authentication required")
	}
	key := tenant.Production()
	if branch != "" {
		var err error
		key, err = tenant.Parse("tadoku/" + branch)
		if err != nil || key == tenant.Production() {
			return nil, ErrIdentityTenantNotServed
		}
	}
	if !a.identityDeployment.Serves(key) {
		return nil, ErrIdentityTenantNotServed
	}
	return a.identity.CheckSession(tenant.WithKey(ctx, key), cookieHeader)
}
