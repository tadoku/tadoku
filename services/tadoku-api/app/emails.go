package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Email = email.Email
type EmailFilter = email.Filter
type EmailStatus = email.Status
type EmailTemplate = mail.Template
type EmailPending = email.Pending
type EmailHeld = email.Held
type EmailSent = email.Sent
type EmailFailed = email.Failed

func (a *Application) ListEmails(ctx context.Context, filter EmailFilter) (email.Page, error) {
	if err := a.requireEmailReader(ctx); err != nil {
		return email.Page{}, err
	}

	var result email.Page
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var err error
		result, err = a.email.List(ctx, filter)
		if err != nil {
			return err
		}
		for _, row := range result.Emails {
			if err := a.auditEmailRead(ctx, row); err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func (a *Application) FindEmail(ctx context.Context, id uuid.UUID) (Email, error) {
	if err := a.requireEmailReader(ctx); err != nil {
		return Email{}, err
	}

	var result Email
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var err error
		result, err = a.email.Find(ctx, id)
		if err != nil {
			return err
		}
		return a.auditEmailRead(ctx, result)
	})
	return result, err
}

func (a *Application) auditEmailRead(ctx context.Context, row Email) error {
	if row.Tenant != tenant.Production() {
		return nil
	}

	metadata := map[string]any{"email_id": row.ID.String()}
	var actor uuid.UUID
	if principal, ok := ctx.Value(emailReaderKey{}).(tokens.Principal); ok && a.hasEmailReader(ctx) {
		actor = principal.TokenID
		metadata["access_token_id"] = principal.TokenID.String()
	} else {
		var err error
		actor, err = identity.RequireActorID(ctx)
		if err != nil {
			return err
		}
	}
	return a.audit.Record(ctx, audit.Event{
		ActorID:  actor,
		Action:   "read_email",
		Metadata: metadata,
	})
}
