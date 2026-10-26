package accountdeletion

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
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
	roles   *authz.Roles
	profile *profile.Service
	audit   *audit.Service
	queue   *jobqueue.Service
}

func NewFiling(roles *authz.Roles, profile *profile.Service, audit *audit.Service, queue *jobqueue.Service) *Filing {
	return &Filing{roles: roles, profile: profile, audit: audit, queue: queue}
}

func (f *Filing) File(ctx context.Context, request Request) (int64, error) {
	return 0, nil
}
