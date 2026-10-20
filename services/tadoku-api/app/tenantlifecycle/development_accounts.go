package tenantlifecycle

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (app *Application) SeedDevelopmentAccounts(
	ctx context.Context,
	key tenant.TestKey,
	fixtures []identity.DevelopmentAccount,
) error {
	return app.registry(ctx, key, false, func(ctx context.Context, _ *queries.Queries) error {
		return identity.SeedDevelopmentAccounts(ctx, identity.NewRepository(app.pool), key, fixtures)
	})
}
