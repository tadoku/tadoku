package api

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (a *Application) IdentityRequestRecovery(ctx context.Context, recovery identity.Recovery) error {
	started := time.Now()
	defer func() {
		timer := time.NewTimer(300*time.Millisecond - time.Since(started))
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
		}
	}()

	request, err := a.identity.AdmitRecovery(ctx, recovery)
	if err != nil || request == nil {
		return err
	}
	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		result, err := a.identity.WriteRecovery(ctx, *request)
		if err != nil {
			return err
		}
		return a.writeIdentityEmails(ctx, result.Emails)
	})
}

func (a *Application) IdentityConsumeRecovery(
	ctx context.Context,
	token identity.RecoveryToken,
	returnTo *string,
) (identity.RecoveryLogin, error) {
	redirect, err := a.identity.RecoverySettingsURL(ctx, returnTo)
	if err != nil {
		return identity.RecoveryLogin{}, err
	}
	var result identity.RecoveryLogin
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var err error
		result, err = a.identity.ConsumeRecovery(ctx, token)
		return err
	})
	result.RedirectTo = redirect
	return result, err
}

func (a *Application) IdentityUpdateProfile(
	ctx context.Context,
	header identity.CookieHeader,
	settings identity.ProfileSettings,
) (identity.AuthenticatedSession, error) {
	session, err := a.identity.SettingsSession(ctx, header)
	if err != nil {
		return identity.AuthenticatedSession{}, err
	}

	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		result, err := a.identity.UpdateProfile(ctx, header, settings)
		if err != nil {
			return err
		}
		if err := a.writeIdentityEmails(ctx, result.Emails); err != nil {
			return err
		}
		session.Account = result.Account
		return nil
	})
	return session, err
}

func (a *Application) IdentityUpdatePassword(ctx context.Context, header identity.CookieHeader, password string) error {
	update, err := a.identity.HashPasswordUpdate(ctx, header, password)
	if err != nil {
		return err
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		return a.identity.ApplyPasswordUpdate(ctx, update)
	})
}

func (a *Application) writeIdentityEmails(ctx context.Context, messages []mail.Message) error {
	for _, message := range messages {
		result, err := a.email.Write(ctx, message)
		if err != nil {
			return err
		}
		if err := a.jobqueue.Enqueue(ctx, result.Jobs...); err != nil {
			return err
		}
	}
	return nil
}
