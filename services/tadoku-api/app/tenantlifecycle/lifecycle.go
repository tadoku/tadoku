package tenantlifecycle

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// Serialize lifecycle commands for one tenant externally, including calls from different processes.
type Application struct {
	pool  *pgxpool.Pool
	keto  *permissions.TenantManager
	cache *leaderboard.Cache
	flipt *fliptmanagement.Client
}

func NewApplication(
	pool *pgxpool.Pool,
	keto *permissions.TenantManager,
	cache *leaderboard.Cache,
	flipt *fliptmanagement.Client,
) *Application {
	return &Application{pool: pool, keto: keto, cache: cache, flipt: flipt}
}

func (app *Application) Provision(
	ctx context.Context,
	key tenant.TestKey,
	testers []uuid.UUID,
	features []fliptmanagement.Resource,
) error {
	for _, tester := range testers {
		if tester == uuid.Nil {
			return errors.New("tester must be a nonzero UUID")
		}
	}
	if err := app.registry(ctx, key, true, func(ctx context.Context, q *queries.Queries) error {
		return q.InsertTestTenant(ctx, key.String())
	}); err != nil {
		return err
	}
	if err := app.keto.Provision(ctx, key, testers); err != nil {
		return fmt.Errorf("provision tenant authorization: %w", err)
	}
	if err := app.flipt.ProvisionTestNamespace(ctx, key, features); err != nil {
		return fmt.Errorf("provision tenant flags: %w", err)
	}
	return nil
}

func (app *Application) Teardown(ctx context.Context, key tenant.TestKey) error {
	if err := app.registry(ctx, key, true, nil); err != nil {
		return err
	}
	if err := app.keto.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete tenant authorization: %w", err)
	}
	if err := app.cache.DeleteTestTenant(ctx, key); err != nil {
		return fmt.Errorf("delete tenant caches: %w", err)
	}
	if err := app.flipt.DeleteTestNamespace(ctx, key); err != nil {
		return fmt.Errorf("delete tenant flags: %w", err)
	}
	return app.registry(ctx, key, true, func(ctx context.Context, q *queries.Queries) error {
		return q.DeleteTestTenant(ctx, key.String())
	})
}

func (app *Application) SetOverride(ctx context.Context, key tenant.TestKey, component string) error {
	if component != jobqueue.WorkerComponent {
		return errors.New("override component is not registered")
	}
	return app.registry(ctx, key, false, func(ctx context.Context, q *queries.Queries) error {
		return q.SetOverride(ctx, queries.SetOverrideParams{Tenant: key.String(), Component: component})
	})
}

func (app *Application) ClearOverride(ctx context.Context, key tenant.TestKey, component string) error {
	if component != jobqueue.WorkerComponent {
		return errors.New("override component is not registered")
	}
	return app.registry(ctx, key, true, func(ctx context.Context, q *queries.Queries) error {
		return q.ClearOverride(ctx, queries.ClearOverrideParams{Tenant: key.String(), Component: component})
	})
}

func (app *Application) registry(
	ctx context.Context,
	key tenant.TestKey,
	allowMissing bool,
	operation func(context.Context, *queries.Queries) error,
) error {
	if key.Key() == (tenant.Key{}) {
		return errors.New("tenant lifecycle requires a parsed test tenant")
	}
	ctx = tenant.WithKey(ctx, key.Key())
	return postgres.RunInTransaction(ctx, app.pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, app.pool)
		if err != nil {
			return err
		}
		q := queries.New(executor)
		owner, err := q.OwnsTenantRegistry(ctx)
		if err != nil {
			return err
		}
		if !owner {
			return errors.New("tenant lifecycle requires database owner credentials")
		}
		if err := q.LockTenant(ctx, key.String()); err != nil {
			return err
		}
		kind, err := q.TenantKind(ctx, key.String())
		if errors.Is(err, pgx.ErrNoRows) {
			if !allowMissing {
				return errors.New("tenant is not provisioned")
			}
		} else if err != nil {
			return err
		} else if kind != "test" {
			return errors.New("tenant lifecycle refuses a production registry row")
		}
		if operation != nil {
			return operation(ctx, q)
		}
		return nil
	})
}
