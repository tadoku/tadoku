package tenantlifecycle

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// SeedDevelopmentAccounts also makes admin an administrator of the test tenant, in the same transaction.
func (app *Application) SeedDevelopmentAccounts(
	ctx context.Context,
	key tenant.TestKey,
	admin identity.DevelopmentAccount,
	reader identity.DevelopmentAccount,
) error {
	return app.registry(ctx, key, false, func(ctx context.Context, _ *queries.Queries) error {
		fixtures := []identity.DevelopmentAccount{admin, reader}
		if err := identity.SeedDevelopmentAccounts(ctx, identity.NewRepository(app.pool), key, fixtures); err != nil {
			return err
		}
		model, err := permissionmodel.New()
		if err != nil {
			return err
		}
		return authz.GrantAdmin(ctx, permissions.NewRelationships(authz.NewRepository(app.pool), model), admin.ID)
	})
}
