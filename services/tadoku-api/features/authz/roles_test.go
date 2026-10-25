package authz

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const granted = errx.Unknown

func TestPermissionsAreValid(t *testing.T) {
	t.Parallel()

	if _, err := permissions.NewModel(Permissions); err != nil {
		t.Error(err)
	}
}

func TestRolesAdministrators(t *testing.T) {
	t.Parallel()
	prodCtx := tenant.WithKey(t.Context(), tenant.Production())

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/roles-0000000a")
	roles, relationships := newTestRoles(t, db.AppPool)
	admin, branchAdmin, user := uuid.New(), uuid.New(), uuid.New()
	grantAdmins(t, prodCtx, relationships, admin)
	grantAdmins(t, branchCtx, relationships, branchAdmin)

	cases := []struct {
		name           string
		ctx            context.Context
		subject        string
		wantAdmin      bool
		wantRequire    errx.Kind
		wantProduction errx.Kind
	}{
		{"production admin in production", prodCtx, admin.String(), true, granted, granted},
		{"user in production", prodCtx, user.String(), false, errx.Forbidden, errx.Forbidden},
		{"production admin on a test tenant", branchCtx, admin.String(), false, errx.Forbidden, errx.Forbidden},
		{"test tenant admin in its tenant", branchCtx, branchAdmin.String(), true, granted, errx.Forbidden},
		{"test tenant admin in production", prodCtx, branchAdmin.String(), false, errx.Forbidden, errx.Forbidden},
		{"nil UUID subject", prodCtx, uuid.Nil.String(), false, errx.Forbidden, errx.Forbidden},
		{"guest", prodCtx, "guest", false, errx.Unauthorized, errx.Unauthorized},
		{"empty subject", prodCtx, "", false, errx.Unauthorized, errx.Unauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := signedIn(tc.ctx, tc.subject)

			isAdmin, err := roles.IsAdmin(ctx)
			orFalse := roles.IsAdminOrFalse(ctx)
			requireErr := roles.RequireAdmin(ctx)
			productionErr := roles.RequireProductionAdmin(ctx)

			if err != nil || isAdmin != tc.wantAdmin {
				t.Errorf("IsAdmin = %t, %v; want %t, nil", isAdmin, err, tc.wantAdmin)
			}
			if orFalse != tc.wantAdmin {
				t.Errorf("IsAdminOrFalse = %t, want %t", orFalse, tc.wantAdmin)
			}
			checkKind(t, "RequireAdmin", requireErr, tc.wantRequire)
			checkKind(t, "RequireProductionAdmin", productionErr, tc.wantProduction)
		})
	}
}

func TestRolesIsProductionAdmin(t *testing.T) {
	t.Parallel()
	prodCtx := tenant.WithKey(t.Context(), tenant.Production())

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/roles-0000000b")
	roles, relationships := newTestRoles(t, db.AppPool)
	admin, branchAdmin := uuid.New(), uuid.New()
	grantAdmins(t, prodCtx, relationships, admin)
	grantAdmins(t, branchCtx, relationships, branchAdmin)

	cases := []struct {
		name    string
		ctx     context.Context
		subject string
		want    bool
	}{
		{"production admin without a tenant", t.Context(), admin.String(), true},
		{"production admin from a test tenant", branchCtx, admin.String(), true},
		{"test tenant admin without a tenant", t.Context(), branchAdmin.String(), false},
		{"test tenant admin from its tenant", branchCtx, branchAdmin.String(), false},
		{"user", t.Context(), uuid.NewString(), false},
		{"guest", t.Context(), "guest", false},
		{"empty subject", t.Context(), "", false},
		{"not a UUID", t.Context(), "not-a-uuid", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := roles.IsProductionAdmin(tc.ctx, tc.subject)

			if err != nil || got != tc.want {
				t.Errorf("IsProductionAdmin = %t, %v; want %t, nil", got, err, tc.want)
			}
		})
	}

	t.Run("storage error", func(t *testing.T) {
		closed, _ := newTestRoles(t, closedPool(t, db))

		got, err := closed.IsProductionAdmin(t.Context(), admin.String())

		if got || errx.KindOf(err) != errx.Unavailable {
			t.Errorf("IsProductionAdmin = %t, %v; want false and unavailable", got, err)
		}
	})
}

func TestRolesAdmit(t *testing.T) {
	t.Parallel()
	prodCtx := tenant.WithKey(t.Context(), tenant.Production())

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/roles-0000000c")
	setup, relationships := newTestRoles(t, db.AppPool)
	bannedUser, branchBanned, unbanned, admin, user := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, userID := range []uuid.UUID{bannedUser, unbanned} {
		if err := setup.SetBanned(prodCtx, userID, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := setup.SetBanned(prodCtx, unbanned, false); err != nil {
		t.Fatal(err)
	}
	if err := setup.SetBanned(branchCtx, branchBanned, true); err != nil {
		t.Fatal(err)
	}
	grantAdmins(t, prodCtx, relationships, admin)
	pool, statements := newLoggedPool(t, db)
	roles, _ := newTestRoles(t, pool)

	cases := []struct {
		name           string
		ctx            context.Context
		subject        string
		want           permissions.Admission
		wantStatements int
	}{
		{"banned user", prodCtx, bannedUser.String(), permissions.Banned{}, 1},
		{"unbanned user", prodCtx, unbanned.String(), permissions.Admitted{}, 1},
		{"user", prodCtx, user.String(), permissions.Admitted{}, 1},
		{"admin", prodCtx, admin.String(), permissions.Admitted{}, 1},
		{"production ban on a test tenant", branchCtx, bannedUser.String(), permissions.Admitted{}, 1},
		{"test tenant ban in its tenant", branchCtx, branchBanned.String(), permissions.Banned{}, 1},
		{"test tenant ban in production", prodCtx, branchBanned.String(), permissions.Admitted{}, 1},
		{"guest", prodCtx, "guest", permissions.Admitted{}, 0},
		{"empty subject", prodCtx, "", permissions.Admitted{}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statements.take()

			got := roles.Admit(signedIn(tc.ctx, tc.subject))

			if got != tc.want {
				t.Errorf("Admit = %#v, want %#v", got, tc.want)
			}
			if sent := statements.take(); len(sent) != tc.wantStatements {
				t.Errorf("sent %d statements, want %d: %q", len(sent), tc.wantStatements, sent)
			}
		})
	}

	t.Run("missing tenant", func(t *testing.T) {
		for _, subject := range []string{user.String(), "guest"} {
			got := roles.Admit(signedIn(t.Context(), subject))

			unavailable, ok := got.(permissions.Unavailable)
			if !ok || errx.KindOf(unavailable.Err) != errx.Unavailable {
				t.Errorf("Admit(%q) = %#v, want Unavailable with an unavailable error", subject, got)
			}
		}
	})
}

func TestRolesFailClosed(t *testing.T) {
	t.Parallel()
	ctx := tenant.WithKey(t.Context(), tenant.Production())

	db := newTestDatabase(t)
	working, relationships := newTestRoles(t, db.AppPool)
	broken, _ := newTestRoles(t, closedPool(t, db))
	admin := uuid.New()
	grantAdmins(t, ctx, relationships, admin)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	lookupErr := errors.New("ban lookup failed")

	cases := []struct {
		name  string
		roles *Roles
		ctx   context.Context
		cause error
	}{
		{"storage error", broken, ctx, nil},
		{"cancelled context", working, cancelled, context.Canceled},
		{"unknown ban", working, permissions.WithBanState(ctx, permissions.BanUnknown(lookupErr)), lookupErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adminCtx := signedIn(tc.ctx, admin.String())

			isAdmin, isAdminErr := tc.roles.IsAdmin(adminCtx)
			requireErr := tc.roles.RequireAdmin(adminCtx)

			if isAdmin || tc.roles.IsAdminOrFalse(adminCtx) {
				t.Error("IsAdmin or IsAdminOrFalse = true, want false")
			}
			for call, err := range map[string]error{"IsAdmin": isAdminErr, "RequireAdmin": requireErr} {
				if errx.KindOf(err) != errx.Unavailable || (tc.cause != nil && !errors.Is(err, tc.cause)) {
					t.Errorf("%s error = %v, want unavailable wrapping %v", call, err, tc.cause)
				}
			}
		})
	}

	t.Run("ban gate", func(t *testing.T) {
		for name, tc := range map[string]struct {
			roles *Roles
			ctx   context.Context
		}{
			"storage error":     {broken, ctx},
			"cancelled context": {working, cancelled},
		} {
			got := tc.roles.Admit(signedIn(tc.ctx, admin.String()))

			if unknown, ok := got.(permissions.AdmittedBanUnknown); !ok || errx.KindOf(unknown.Err) != errx.Unavailable {
				t.Errorf("%s: Admit = %#v, want AdmittedBanUnknown with an unavailable error", name, got)
			}
		}
	})
}

func TestRolesUserRoles(t *testing.T) {
	t.Parallel()
	ctx := tenant.WithKey(t.Context(), tenant.Production())

	db := newTestDatabase(t)
	setup, relationships := newTestRoles(t, db.AppPool)
	admin, bannedUser, both, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	grantAdmins(t, ctx, relationships, admin, both)
	for _, userID := range []uuid.UUID{bannedUser, both} {
		if err := setup.SetBanned(ctx, userID, true); err != nil {
			t.Fatal(err)
		}
	}
	pool, statements := newLoggedPool(t, db)
	roles, _ := newTestRoles(t, pool)

	got, err := roles.UserRoles(ctx, []uuid.UUID{admin, bannedUser, both, stranger})

	if err != nil {
		t.Fatal(err)
	}
	want := map[uuid.UUID]Role{admin: RoleAdmin, bannedUser: RoleBanned, both: RoleAdmin, stranger: RoleUser}
	if !maps.Equal(got, want) {
		t.Errorf("UserRoles = %v, want %v", got, want)
	}
	if sent := statements.take(); len(sent) != 1 {
		t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
	}

	_, oneErr := roles.UserRoles(t.Context(), []uuid.UUID{admin})
	_, noneErr := roles.UserRoles(t.Context(), nil)
	for _, err := range []error{oneErr, noneErr} {
		if errx.KindOf(err) != errx.Unavailable {
			t.Errorf("UserRoles without a tenant error = %v, want unavailable", err)
		}
	}
}

func newTestRoles(t *testing.T, pool *pgxpool.Pool) (*Roles, *permissions.Relationships) {
	t.Helper()

	model, err := permissions.NewModel(Permissions)
	if err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(pool)
	relationships := permissions.NewRelationships(repository, model)
	return NewRoles(permissions.NewChecker(repository, model), relationships), relationships
}

func grantAdmins(t *testing.T, ctx context.Context, relationships *permissions.Relationships, userIDs ...uuid.UUID) {
	t.Helper()

	for _, userID := range userIDs {
		if err := GrantAdmin(ctx, relationships, userID); err != nil {
			t.Fatal(err)
		}
	}
}

func signedIn(ctx context.Context, subject string) context.Context {
	return requestidentity.WithUser(ctx, &requestidentity.User{Subject: subject})
}

func checkKind(t *testing.T, call string, err error, want errx.Kind) {
	t.Helper()

	if errx.KindOf(err) != want || (err == nil) != (want == granted) {
		t.Errorf("%s error = %v, want kind %v (nil when granted)", call, err, want)
	}
}
