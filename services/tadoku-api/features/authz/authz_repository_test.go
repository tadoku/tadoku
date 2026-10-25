package authz

import (
	"context"
	"errors"
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
		if err := repository.Insert(ctx, []permissions.Tuple{admin}); err != nil {
			return err
		}

		subjects, err := repository.Subjects(ctx, []permissions.SubjectsKey{{
			ObjectType: admin.ObjectType,
			ObjectID:   admin.ObjectID,
			Relation:   admin.Relation,
			UserID:     admin.SubjectID,
		}})
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

	if stored := storedRelationships(t, db, tenant.Production().String()); len(stored) != 0 {
		t.Errorf("stored after rollback = %q, want none", stored)
	}
}

func TestRepositoryRequiresATenant(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	admin := userTuple(uuid.New())

	_, subjectsErr := repository.Subjects(t.Context(), nil)
	errs := []error{
		subjectsErr,
		repository.Insert(t.Context(), []permissions.Tuple{admin}),
		repository.Insert(t.Context(), nil),
		repository.Delete(t.Context(), []permissions.Tuple{admin}),
		repository.Delete(t.Context(), nil),
	}
	for _, err := range errs {
		if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
			t.Errorf("error = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
		}
	}
}

func TestRelationshipStorage(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	userID, contestID, groupID, logID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, testModel),
		siteAdminRelationship(userID),
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
