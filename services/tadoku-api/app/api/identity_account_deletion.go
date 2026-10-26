package api

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (a *Application) IdentityRequestAccountDeletion(
	ctx context.Context,
	header identity.CookieHeader,
	email string,
) (identity.Cookie, error) {
	deletion, err := a.identity.AdmitAccountDeletion(ctx, header, email)
	if err != nil {
		return identity.Cookie{}, err
	}

	var cookie identity.Cookie
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		userID, cleared, err := a.identity.RequestAccountDeletion(ctx, deletion)
		if err != nil {
			return err
		}
		if err := a.fileAccountDeletion(ctx, SelfServiceAccountDeletion(userID)); err != nil {
			return err
		}
		cookie = cleared
		return nil
	})
	return cookie, err
}
