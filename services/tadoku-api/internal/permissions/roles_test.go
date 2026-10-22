package permissions

import (
	"errors"
	"maps"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestRoles(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	pool, statements := newLoggedAppPool(t, db)
	roles := NewRoles(pool)
	admin, bannedUser, both, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	grantAll(t, ctx, NewGrants(db.AppPool), Admin(admin), Ban(bannedUser), Admin(both), Ban(both))

	t.Run("batch", func(t *testing.T) {
		subjects := []string{
			admin.String(),
			bannedUser.String(),
			both.String(),
			stranger.String(),
			"",
			"guest",
			"not-a-uuid",
		}

		got, err := roles.RolesForSubjects(ctx, subjects)

		if err != nil {
			t.Fatal(err)
		}
		want := map[string]TargetRoles{
			admin.String():      {Admin: true},
			bannedUser.String(): {Banned: true},
			both.String():       {Admin: true, Banned: true},
			stranger.String():   {},
			"":                  {},
			"guest":             {},
			"not-a-uuid":        {},
		}
		if !maps.Equal(got, want) {
			t.Errorf("roles = %v, want %v", got, want)
		}
		if sent := statements.take(); len(sent) != 1 {
			t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
		}
	})

	t.Run("one subject", func(t *testing.T) {
		got, err := roles.RolesForSubject(ctx, both.String())

		if err != nil {
			t.Fatal(err)
		}
		if want := (TargetRoles{Admin: true, Banned: true}); got != want {
			t.Errorf("roles = %+v, want %+v", got, want)
		}
		if sent := statements.take(); len(sent) != 1 {
			t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
		}
	})

	t.Run("subjects that cannot be stored", func(t *testing.T) {
		for _, subject := range []string{"", "guest", "not-a-uuid"} {
			got, err := roles.RolesForSubject(ctx, subject)

			if err != nil || got != (TargetRoles{}) {
				t.Errorf("roles for %q = %+v, %v; want no roles", subject, got, err)
			}
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		_, oneErr := roles.RolesForSubject(t.Context(), admin.String())
		_, batchErr := roles.RolesForSubjects(t.Context(), []string{"guest"})

		for _, err := range []error{oneErr, batchErr} {
			if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
				t.Errorf("error = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
			}
		}
	})
}
