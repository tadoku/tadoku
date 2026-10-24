package permissions

import (
	"maps"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestRelationshipsWrite(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	writer := NewRelationships(db.AppPool, testModel)
	production := tenant.Production().String()
	twice := uuid.New()

	grantAll(t, ctx, writer, adminRelationship(twice), adminRelationship(twice))
	want := []string{"site:root#admins@user:" + twice.String() + "#"}
	if stored := storedRelationships(t, db, production); !slices.Equal(stored, want) {
		t.Errorf("stored after granting twice = %q, want %q", stored, want)
	}

	if err := writer.Revoke(ctx, banRelationship(twice)); err != nil {
		t.Errorf("revoke a missing relationship: %v", err)
	}
	if err := writer.Revoke(ctx, adminRelationship(twice)); err != nil {
		t.Errorf("revoke a stored relationship: %v", err)
	}
	if stored := storedRelationships(t, db, production); len(stored) != 0 {
		t.Errorf("stored after revoke = %q, want none", stored)
	}

	grantErr := writer.Grant(t.Context(), adminRelationship(twice))
	revokeErr := writer.Revoke(t.Context(), adminRelationship(twice))
	for _, err := range []error{grantErr, revokeErr} {
		if errx.KindOf(err) != errx.Unavailable {
			t.Errorf("write without tenant error = %v, want unavailable", err)
		}
	}
}

func TestStoredRelationships(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	pool, statements := newLoggedAppPool(t, db)
	writer := NewRelationships(pool, testModel)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000c")
	admin, banned, both, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contestID, groupID := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		writer,
		adminRelationship(admin),
		banRelationship(banned),
		adminRelationship(both),
		banRelationship(both),
		contestModeratorGroup(contestID, groupID),
	)
	grantAll(t, branchCtx, writer, adminRelationship(stranger))
	statements.take()

	t.Run("batch", func(t *testing.T) {
		candidates := []Relationship{
			adminRelationship(admin),
			banRelationship(admin),
			banRelationship(banned),
			adminRelationship(both),
			banRelationship(both),
			adminRelationship(stranger),
			contestModeratorGroup(contestID, groupID),
			contestModeratorGroup(contestID, uuid.New()),
		}

		got, err := writer.Stored(ctx, candidates)

		if err != nil {
			t.Fatal(err)
		}
		want := map[Relationship]bool{
			adminRelationship(admin):                  true,
			banRelationship(banned):                   true,
			adminRelationship(both):                   true,
			banRelationship(both):                     true,
			contestModeratorGroup(contestID, groupID): true,
		}
		if !maps.Equal(got, want) {
			t.Errorf("stored = %v, want %v", got, want)
		}
		if sent := statements.take(); len(sent) != 1 {
			t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
		}
	})

	t.Run("no candidates", func(t *testing.T) {
		got, err := writer.Stored(ctx, nil)

		if err != nil || len(got) != 0 {
			t.Errorf("stored = %v, %v; want nothing", got, err)
		}
		if sent := statements.take(); len(sent) != 0 {
			t.Errorf("sent %d statements, want none: %q", len(sent), sent)
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		_, oneErr := writer.Stored(t.Context(), []Relationship{adminRelationship(admin)})
		_, noneErr := writer.Stored(t.Context(), nil)

		for _, err := range []error{oneErr, noneErr} {
			if errx.KindOf(err) != errx.Unavailable {
				t.Errorf("error = %v, want unavailable", err)
			}
		}
	})
}

func TestRelationshipStorage(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	userID, contestID, groupID, logID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	grantAll(
		t,
		ctx,
		NewRelationships(db.AppPool, testModel),
		adminRelationship(userID),
		contestModeratorGroup(contestID, groupID),
		logCountsFor(logID, contestID),
	)

	want := []string{
		"contest:" + contestID.String() + "#moderators@group:" + groupID.String() + "#members",
		"log:" + logID.String() + "#contests@contest:" + contestID.String() + "#",
		"site:root#admins@user:" + userID.String() + "#",
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
	writer := NewRelationships(db.AppPool, testModel)
	userID := uuid.New()
	grantAll(t, branchCtx, writer, adminRelationship(userID), banRelationship(userID))
	grantAll(t, prodCtx, writer, adminRelationship(userID))
	if stored := storedRelationships(t, db, branch.String()); len(stored) != 2 {
		t.Fatalf("stored in test tenant = %q, want two rows", stored)
	}

	if _, err := db.Pool.Exec(t.Context(), `delete from tenants where key = $1`, branch.String()); err != nil {
		t.Fatal(err)
	}

	if stored := storedRelationships(t, db, branch.String()); len(stored) != 0 {
		t.Errorf("stored in deleted tenant = %q, want none", stored)
	}
	want := []string{"site:root#admins@user:" + userID.String() + "#"}
	if stored := storedRelationships(t, db, tenant.Production().String()); !slices.Equal(stored, want) {
		t.Errorf("stored in production = %q, want %q", stored, want)
	}
}
