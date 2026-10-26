package accountdeletion

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type source string

const (
	operator    source = "operator"
	selfService source = "self_service"
)

// Request is a deletion request for one user. Build it with OperatorRequest or SelfServiceRequest.
type Request struct {
	userID  uuid.UUID
	actorID uuid.UUID
	source  source
	reason  string
}

func OperatorRequest(userID, operatorID uuid.UUID, reason string) (Request, error) {
	if userID == uuid.Nil {
		return Request{}, errx.NewInvalidInputError("user id is required")
	}
	if operatorID == uuid.Nil {
		return Request{}, errx.NewInvalidInputError("operator user id is required")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return Request{}, errx.NewInvalidInputError("reason is required")
	}
	return Request{userID: userID, actorID: operatorID, source: operator, reason: reason}, nil
}

func SelfServiceRequest(userID uuid.UUID) Request {
	return Request{userID: userID, actorID: userID, source: selfService}
}

type Filing struct {
	roles    *authz.Roles
	accounts *identity.Deletion
	profile  *profile.Service
	audit    *audit.Service
	queue    *jobqueue.Service
}

func NewFiling(
	roles *authz.Roles,
	accounts *identity.Deletion,
	profile *profile.Service,
	audit *audit.Service,
	queue *jobqueue.Service,
) *Filing {
	return &Filing{roles: roles, accounts: accounts, profile: profile, audit: audit, queue: queue}
}

// File requires the caller's transaction and returns the ID of the account.delete.v1 job. It refuses an administrator,
// and a banned user only when the user asks; from its commit the user can no longer sign in or change data.
func (f *Filing) File(ctx context.Context, request Request) (int64, error) {
	roles, err := f.roles.UserRoles(ctx, []uuid.UUID{request.userID})
	if err != nil {
		return 0, err
	}
	switch roles[request.userID] {
	case authz.RoleAdmin:
		return 0, identity.ErrAdminAccount
	case authz.RoleBanned:
		if request.source == selfService {
			return 0, identity.ErrBannedAccount
		}
	}

	if err := f.accounts.LockAccount(ctx, request.userID); err != nil {
		return 0, err
	}
	if err := f.profile.LockForDeletion(ctx, request.userID, timex.Now()); err != nil {
		return 0, err
	}
	err = f.audit.Record(ctx, audit.Event{
		ActorID: request.actorID,
		Action:  "request_account_deletion",
		Metadata: map[string]any{
			"target_user_id": request.userID.String(),
			"source":         string(request.source),
		},
		Description: request.reason,
	})
	if err != nil {
		return 0, err
	}
	return f.queue.EnqueueOne(ctx, jobs.DeleteAccountV1{UserID: request.userID})
}
