package tenantlifecycle

import (
	"context"
	"errors"

	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// SeedDevelopmentAccounts also makes admin an administrator of the test tenant, in the same transaction and under
// admin's users row lock. It returns authz.ErrBannedUser for a banned admin, and grants nothing to an admin whose
// account is being deleted.
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

		users := profile.NewService(profile.NewRepository(app.pool), identity.NewReader(app.pool))
		err := users.LockUser(ctx, admin.ID)
		if errors.Is(err, profile.ErrAccountDeletionInProgress) {
			return nil
		}
		if err != nil {
			return err
		}

		model, err := permissionmodel.New()
		if err != nil {
			return err
		}
		return authz.GrantAdmin(ctx, permissions.NewRelationships(authz.NewRepository(app.pool), model), admin.ID)
	})
}
