package relationships

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestStoreJoinsTheContextTransaction(t *testing.T) {
	t.Parallel()
	ctx := tenant.WithKey(t.Context(), tenant.Production())

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	store := NewStore(db.AppPool)
	admin := userTuple(uuid.New())

	errRollback := errors.New("roll back")
	err = postgres.RunInTransaction(ctx, db.AppPool, func(ctx context.Context) error {
		if err := store.Insert(ctx, admin); err != nil {
			return err
		}

		subjects, err := store.Subjects(ctx, admin.Object, admin.Relation, admin.Subject.ID)
		if err != nil {
			return err
		}
		if len(subjects) != 1 || subjects[0] != admin.Subject {
			t.Errorf("subjects inside the transaction = %+v, want %+v", subjects, admin.Subject)
		}
		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatalf("transaction error = %v, want %v", err, errRollback)
	}

	stored, err := store.Stored(ctx, []Tuple{admin})
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 0 {
		t.Errorf("stored after rollback = %+v, want none", stored)
	}
}

func TestStoreRequiresATenant(t *testing.T) {
	t.Parallel()

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	store := NewStore(db.AppPool)
	admin := userTuple(uuid.New())

	_, subjectsErr := store.Subjects(t.Context(), admin.Object, admin.Relation, admin.Subject.ID)
	_, storedErr := store.Stored(t.Context(), nil)
	errs := []error{
		subjectsErr,
		store.Insert(t.Context(), admin),
		store.Delete(t.Context(), admin),
		storedErr,
	}
	for _, err := range errs {
		if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
			t.Errorf("error = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
		}
	}
}

func userTuple(userID uuid.UUID) Tuple {
	return Tuple{
		Object:   Object{Type: "site", ID: "root"},
		Relation: "admins",
		Subject:  Subject{Type: UserType, ID: userID.String()},
	}
}
