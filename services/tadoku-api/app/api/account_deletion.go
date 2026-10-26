package api

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type accountDeletionSource string

const (
	adminDeletion       accountDeletionSource = "admin"
	selfServiceDeletion accountDeletionSource = "self_service"
)

// AccountDeletionParameters is an administrator's request to delete a user's account. Email is the account's address
// as the administrator typed it to confirm the target.
type AccountDeletionParameters struct {
	UserID uuid.UUID
	Email  string
	Reason string
}

// AccountDeletionRequest is a deletion request for one user. Build it with AdminAccountDeletion or
// SelfServiceAccountDeletion.
type AccountDeletionRequest struct {
	userID  uuid.UUID
	actorID uuid.UUID
	source  accountDeletionSource
	email   string
	reason  string
}

// AdminAccountDeletion requires the target account's email address, which filing compares with the account's
// address, and a reason for the audit log.
func AdminAccountDeletion(userID, adminID uuid.UUID, email, reason string) (AccountDeletionRequest, error) {
	if userID == uuid.Nil {
		return AccountDeletionRequest{}, errx.NewInvalidInputError("user id is required")
	}
	if adminID == uuid.Nil {
		return AccountDeletionRequest{}, errx.NewInvalidInputError("administrator user id is required")
	}
	if strings.TrimSpace(email) == "" {
		return AccountDeletionRequest{}, errx.NewInvalidInputError("email is required")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return AccountDeletionRequest{}, errx.NewInvalidInputError("reason is required")
	}
	if utf8.RuneCountInString(reason) > 1000 {
		return AccountDeletionRequest{}, errx.NewInvalidInputError("reason must be at most 1000 characters")
	}
	return AccountDeletionRequest{
		userID:  userID,
		actorID: adminID,
		source:  adminDeletion,
		email:   email,
		reason:  reason,
	}, nil
}

func SelfServiceAccountDeletion(userID uuid.UUID) AccountDeletionRequest {
	return AccountDeletionRequest{userID: userID, actorID: userID, source: selfServiceDeletion}
}

// RequestAccountDeletion files an administrator's deletion request for another user, with the administrator as the
// actor.
func (a *Application) RequestAccountDeletion(ctx context.Context, parameters AccountDeletionParameters) error {
	if err := a.roles.RequireAdmin(ctx); err != nil {
		return err
	}

	adminID, err := requestidentity.RequireActorID(ctx)
	if err != nil {
		return err
	}
	request, err := AdminAccountDeletion(parameters.UserID, adminID, parameters.Email, parameters.Reason)
	if err != nil {
		return err
	}
	return a.FileAccountDeletion(ctx, request)
}

// FileAccountDeletion files the request in its own transaction and queues account.delete.v1. It checks no actor:
// call it from an operation that has authorized the actor, such as RequestAccountDeletion. An unknown or deleted user
// is not found. An administrator's request must name the account's address. It refuses an administrator, and a banned
// user only when the user asks, and a refusal writes nothing. It reads roles once it holds the users row, which role
// changes lock too, so a ban or grant that commits first counts. From its commit the user can no longer sign in or
// change data.
func (a *Application) FileAccountDeletion(ctx context.Context, request AccountDeletionRequest) error {
	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		return a.fileAccountDeletion(ctx, request)
	})
}

// fileAccountDeletion is FileAccountDeletion inside the caller's transaction. A caller that locks the account row
// first, as a self-service request does, keeps the account-then-user lock order.
func (a *Application) fileAccountDeletion(ctx context.Context, request AccountDeletionRequest) error {
	if err := a.identityDeletion.LockAccountForRequest(ctx, request.userID); err != nil {
		return err
	}
	if err := a.profile.LockForDeletion(ctx, request.userID, timex.Now()); err != nil {
		return err
	}
	if request.source == adminDeletion {
		if err := a.identityDeletion.ConfirmAccount(ctx, request.userID, request.email); err != nil {
			return err
		}
	}

	roles, err := a.roles.UserRoles(ctx, []uuid.UUID{request.userID})
	if err != nil {
		return err
	}
	switch roles[request.userID] {
	case authz.RoleAdmin:
		return identity.ErrAdminAccount
	case authz.RoleBanned:
		if request.source == selfServiceDeletion {
			return identity.ErrBannedAccount
		}
	}

	err = a.audit.Record(ctx, audit.Event{
		ActorID: request.actorID,
		Action:  "request_account_deletion",
		Metadata: map[string]any{
			"target_user_id": request.userID.String(),
			"source":         string(request.source),
		},
		Description: request.reason,
	})
	if err != nil {
		return err
	}
	return a.jobqueue.Enqueue(ctx, jobs.DeleteAccountV1{UserID: request.userID})
}
