package permissions

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestGrantsWrite(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	grants := NewGrants(db.AppPool)
	production := tenant.Production().String()
	rolledBack, twice := uuid.New(), uuid.New()

	errRollback := errors.New("roll back")
	err := postgres.RunInTransaction(ctx, db.AppPool, func(ctx context.Context) error {
		if err := grants.Grant(ctx, Admin(rolledBack)); err != nil {
			return err
		}

		allowed, err := model.check(ctx, store{db: db.AppPool}, appAdmin, rolledBack)
		if err != nil {
			return err
		}
		if !allowed {
			t.Error("grant is not visible inside its transaction")
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("transaction error = %v, want %v", err, errRollback)
	}
	if stored := storedRelationships(t, db, production); len(stored) != 0 {
		t.Errorf("stored after rollback = %q, want none", stored)
	}

	grantAll(t, ctx, grants, Admin(twice), Admin(twice))
	want := []string{"app:root#admins@user:" + twice.String() + "#"}
	if stored := storedRelationships(t, db, production); !slices.Equal(stored, want) {
		t.Errorf("stored after granting twice = %q, want %q", stored, want)
	}

	if err := grants.Revoke(ctx, Ban(twice)); err != nil {
		t.Errorf("revoke a missing relationship: %v", err)
	}
	if err := grants.Revoke(ctx, Admin(twice)); err != nil {
		t.Errorf("revoke a stored relationship: %v", err)
	}
	if stored := storedRelationships(t, db, production); len(stored) != 0 {
		t.Errorf("stored after revoke = %q, want none", stored)
	}

	grantErr := grants.Grant(t.Context(), Admin(twice))
	revokeErr := grants.Revoke(t.Context(), Admin(twice))
	for _, err := range []error{grantErr, revokeErr} {
		if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
			t.Errorf("write without tenant error = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
		}
	}
}

func TestRelationshipStorage(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	userID, contestID, groupID, logID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	grantAll(
		t,
		ctx,
		&Grants{model: testModel, store: store{db: db.AppPool}},
		Admin(userID),
		contestModeratorGroup(contestID, groupID),
		logCountsFor(logID, contestID),
	)

	want := []string{
		"app:root#admins@user:" + userID.String() + "#",
		"contest:" + contestID.String() + "#moderators@group:" + groupID.String() + "#members",
		"log:" + logID.String() + "#contests@contest:" + contestID.String() + "#",
	}
	if stored := storedRelationships(t, db, tenant.Production().String()); !slices.Equal(stored, want) {
		t.Errorf("stored = %q, want %q", stored, want)
	}
}

func TestTenantDeletionRemovesRelationships(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000b")
	branch, _ := tenant.FromContext(branchCtx)
	grants := NewGrants(db.AppPool)
	userID := uuid.New()
	grantAll(t, branchCtx, grants, Admin(userID), Ban(userID))
	grantAll(t, prodCtx, grants, Admin(userID))
	if stored := storedRelationships(t, db, branch.String()); len(stored) != 2 {
		t.Fatalf("stored in test tenant = %q, want two rows", stored)
	}

	if _, err := db.Pool.Exec(t.Context(), `delete from tenants where key = $1`, branch.String()); err != nil {
		t.Fatal(err)
	}

	if stored := storedRelationships(t, db, branch.String()); len(stored) != 0 {
		t.Errorf("stored in deleted tenant = %q, want none", stored)
	}
	want := []string{"app:root#admins@user:" + userID.String() + "#"}
	if stored := storedRelationships(t, db, tenant.Production().String()); !slices.Equal(stored, want) {
		t.Errorf("stored in production = %q, want %q", stored, want)
	}
}
