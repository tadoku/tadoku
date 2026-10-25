package authz

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestRepositoryJoinsTheContextTransaction(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	admin := userTuple(uuid.New())

	errRollback := errors.New("roll back")
	err := postgres.RunInTransaction(ctx, db.AppPool, func(ctx context.Context) error {
		if err := repository.Insert(ctx, admin); err != nil {
			return err
		}

		subjects, err := repository.Subjects(ctx, admin.ObjectType, admin.ObjectID, admin.Relation, admin.SubjectID)
		if err != nil {
			return err
		}
		if len(subjects) != 1 || subjects[0] != admin {
			t.Errorf("subjects inside the transaction = %+v, want %+v", subjects, admin)
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("transaction error = %v, want %v", err, errRollback)
	}

	stored, err := repository.Stored(ctx, []permissions.Tuple{admin})
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 0 {
		t.Errorf("stored after rollback = %+v, want none", stored)
	}
}

func TestRepositoryRequiresATenant(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	admin := userTuple(uuid.New())

	_, subjectsErr := repository.Subjects(
		t.Context(),
		admin.ObjectType,
		admin.ObjectID,
		admin.Relation,
		admin.SubjectID,
	)
	_, storedErr := repository.Stored(t.Context(), nil)
	errs := []error{
		subjectsErr,
		repository.Insert(t.Context(), admin),
		repository.Delete(t.Context(), admin),
		storedErr,
	}
	for _, err := range errs {
		if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
			t.Errorf("error = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
		}
	}
}

func TestRelationshipsWrite(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	writer := permissions.NewRelationships(NewRepository(db.AppPool), testModel)
	production := tenant.Production().String()
	twice := uuid.New()

	grantAll(t, ctx, writer, siteAdminRelationship(twice), siteAdminRelationship(twice))
	want := []string{"site:root#admins@user:" + twice.String() + "#"}
	if stored := storedRelationships(t, db, production); !slices.Equal(stored, want) {
		t.Errorf("stored after granting twice = %q, want %q", stored, want)
	}

	if err := writer.Revoke(ctx, siteBanRelationship(twice)); err != nil {
		t.Errorf("revoke a missing relationship: %v", err)
	}
	if err := writer.Revoke(ctx, siteAdminRelationship(twice)); err != nil {
		t.Errorf("revoke a stored relationship: %v", err)
	}
	if stored := storedRelationships(t, db, production); len(stored) != 0 {
		t.Errorf("stored after revoke = %q, want none", stored)
	}

	grantErr := writer.Grant(t.Context(), siteAdminRelationship(twice))
	revokeErr := writer.Revoke(t.Context(), siteAdminRelationship(twice))
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
	pool, statements := newLoggedPool(t, db)
	repository := NewRepository(pool)
	writer := permissions.NewRelationships(repository, testModel)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000c")
	admin, banned, both, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contestID, groupID := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		writer,
		siteAdminRelationship(admin),
		siteBanRelationship(banned),
		siteAdminRelationship(both),
		siteBanRelationship(both),
	)
	insertAll(t, ctx, repository, contestModeratorGroup(contestID, groupID))
	grantAll(t, branchCtx, writer, siteAdminRelationship(stranger))
	statements.take()

	t.Run("batch", func(t *testing.T) {
		candidates := []permissions.Relationship{
			siteAdminRelationship(admin),
			siteBanRelationship(admin),
			siteBanRelationship(banned),
			siteAdminRelationship(both),
			siteBanRelationship(both),
			siteAdminRelationship(stranger),
		}

		got, err := writer.Stored(ctx, candidates)

		if err != nil {
			t.Fatal(err)
		}
		want := map[permissions.Relationship]bool{
			siteAdminRelationship(admin): true,
			siteBanRelationship(banned):  true,
			siteAdminRelationship(both):  true,
			siteBanRelationship(both):    true,
		}
		if !maps.Equal(got, want) {
			t.Errorf("stored = %v, want %v", got, want)
		}
		if sent := statements.take(); len(sent) != 1 {
			t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
		}
	})

	t.Run("subject sets", func(t *testing.T) {
		stored := contestModeratorGroup(contestID, groupID)

		got, err := repository.Stored(ctx, []permissions.Tuple{stored, contestModeratorGroup(contestID, uuid.New())})

		if err != nil {
			t.Fatal(err)
		}
		if want := []permissions.Tuple{stored}; !slices.Equal(got, want) {
			t.Errorf("stored = %+v, want %+v", got, want)
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
		_, oneErr := writer.Stored(t.Context(), []permissions.Relationship{siteAdminRelationship(admin)})
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
	repository := NewRepository(db.AppPool)
	userID, contestID, groupID, logID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	grantAll(t, ctx, permissions.NewRelationships(repository, testModel), siteAdminRelationship(userID))
	insertAll(t, ctx, repository, contestModeratorGroup(contestID, groupID), logCountsFor(logID, contestID))

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
	writer := permissions.NewRelationships(NewRepository(db.AppPool), testModel)
	userID := uuid.New()
	grantAll(t, branchCtx, writer, siteAdminRelationship(userID), siteBanRelationship(userID))
	grantAll(t, prodCtx, writer, siteAdminRelationship(userID))
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

func userTuple(userID uuid.UUID) permissions.Tuple {
	return permissions.Tuple{
		ObjectType:  string(siteType),
		ObjectID:    siteRoot,
		Relation:    "admins",
		SubjectType: permissions.UserSubjectType,
		SubjectID:   userID.String(),
	}
}
