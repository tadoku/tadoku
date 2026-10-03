package jobqueue

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant/alltenants"
)

var (
	ErrNotFailed = errors.New("job is not failed or supported")
)

const WorkerComponent = "tadoku-worker"

type ClaimedJob struct {
	ID             int64
	Tenant         tenant.Key
	Type           jobs.Type
	Payload        json.RawMessage
	Attempts       int
	Token          uuid.UUID
	LeaseExpiresAt time.Time
	Reclaimed      bool
}

type Stats struct {
	Pending     int64
	Failed      int64
	OldestDueAt *time.Time
}

type UnsupportedStats struct {
	Pending     int64
	Running     int64
	Failed      int64
	OldestDueAt *time.Time
}

func validateErrorCode(code string) error {
	if len(code) == 0 || len(code) > 100 {
		return errx.NewInvalidInputError("job error code must be 1 to 100 characters")
	}
	for _, r := range code {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return errx.NewInvalidInputError("job error code must contain only lowercase letters, digits, or underscores")
		}
	}
	return nil
}

var componentPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type Scope struct {
	key       tenant.Key
	component string
}

func AllTenantsExcept(component string) (Scope, error) {
	if !componentPattern.MatchString(component) {
		return Scope{}, errors.New(
			"worker component must contain lowercase letters, digits or hyphens and start with a letter",
		)
	}
	return Scope{component: component}, nil
}

func OnlyTenant(key tenant.Key) (Scope, error) {
	if key == (tenant.Key{}) || key == tenant.Production() {
		return Scope{}, errors.New("branch worker requires a parsed non-production tenant")
	}
	return Scope{key: key}, nil
}

type queueScopeKey struct{}

// Context scopes queue claims, backlog and retention; handlers need the claimed job's tenant instead.
func (scope Scope) Context(ctx context.Context) (context.Context, error) {
	if scope == (Scope{}) {
		return nil, errors.New("worker requires a queue scope")
	}

	ctx = context.WithValue(ctx, queueScopeKey{}, scope)
	if scope.key != (tenant.Key{}) {
		return tenant.WithKey(ctx, scope.key), nil
	}
	return alltenants.With(ctx), nil
}

func componentFromContext(ctx context.Context) *string {
	scope, _ := ctx.Value(queueScopeKey{}).(Scope)
	if scope.component == "" {
		return nil
	}
	return &scope.component
}
