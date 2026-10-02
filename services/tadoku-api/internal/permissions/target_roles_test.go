package permissions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

type fakeKeto struct {
	results         map[string]ketoclient.PermissionResult
	subjectIDsByRel map[string][]string
	listSubjectsErr error
}

func (f *fakeKeto) CheckPermission(
	ctx context.Context,
	namespace, object, relation string,
	subject ketoclient.Subject,
) (bool, error) {
	r, ok := f.results[relation]
	if !ok {
		return false, errors.New("missing relation in fake")
	}
	return r.Allowed, r.Err
}

func (f *fakeKeto) CheckPermissions(ctx context.Context, checks []ketoclient.PermissionCheck) []ketoclient.PermissionResult {
	out := make([]ketoclient.PermissionResult, 0, len(checks))
	for _, c := range checks {
		r, ok := f.results[c.Relation]
		if !ok {
			out = append(out, ketoclient.PermissionResult{
				Check:   c,
				Allowed: false,
				Err:     errors.New("missing relation in fake"),
			})
			continue
		}
		out = append(out, ketoclient.PermissionResult{Check: c, Allowed: r.Allowed, Err: r.Err})
	}
	return out
}

func (f *fakeKeto) ListSubjectIDsForRelation(ctx context.Context, namespace, object, relation string) ([]string, error) {
	if f.listSubjectsErr != nil {
		return nil, f.listSubjectsErr
	}
	return f.subjectIDsByRel[relation], nil
}

func TestKetoService_RolesForSubject_Guest(t *testing.T) {
	svc := NewKetoService(&fakeKeto{})
	claims, err := svc.RolesForSubject(tenant.WithKey(t.Context(), tenant.Production()), "guest")
	require.NoError(t, err)
	assert.Equal(t, TargetRoles{}, claims)
}

func TestKetoService_RolesForSubject_Admin(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admin":     {Allowed: true},
			"is_banned": {Allowed: false},
		},
	})

	claims, err := svc.RolesForSubject(tenant.WithKey(t.Context(), tenant.Production()), "kratos-id")
	require.NoError(t, err)
	assert.True(t, claims.Admin)
	assert.False(t, claims.Banned)
}

func TestKetoService_RolesForSubject_Banned(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admin":     {Allowed: false},
			"is_banned": {Allowed: true},
		},
	})

	claims, err := svc.RolesForSubject(tenant.WithKey(t.Context(), tenant.Production()), "kratos-id")
	require.NoError(t, err)
	assert.False(t, claims.Admin)
	assert.True(t, claims.Banned)
}

func TestKetoService_RolesForSubject_Error(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		results: map[string]ketoclient.PermissionResult{
			"admin":     {Allowed: false, Err: errors.New("boom")},
			"is_banned": {Allowed: false},
		},
	})

	_, err := svc.RolesForSubject(tenant.WithKey(t.Context(), tenant.Production()), "kratos-id")
	require.Error(t, err)
}

func TestKetoService_RolesForSubjects(t *testing.T) {
	svc := NewKetoService(&fakeKeto{
		subjectIDsByRel: map[string][]string{
			"admins": {"a"},
			"banned": {"b"},
		},
	})

	claimsBySubject, err := svc.RolesForSubjects(
		tenant.WithKey(t.Context(), tenant.Production()),
		[]string{"a", "b", "c", "guest", ""},
	)
	require.NoError(t, err)

	assert.True(t, claimsBySubject["a"].Admin)
	assert.False(t, claimsBySubject["a"].Banned)

	assert.False(t, claimsBySubject["b"].Admin)
	assert.True(t, claimsBySubject["b"].Banned)

	assert.False(t, claimsBySubject["c"].Admin)
	assert.False(t, claimsBySubject["c"].Banned)
	assert.Equal(t, TargetRoles{}, claimsBySubject["guest"])
	assert.Equal(t, TargetRoles{}, claimsBySubject[""])

}

func TestTargetRolesFollowTenant(t *testing.T) {
	fixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := fixture.Reset(t.Context(), "testdata/tenant_relationships.json"); err != nil {
		t.Fatal(err)
	}

	key, err := tenant.Parse("e2e/alpha-0000000a")
	if err != nil {
		t.Fatal(err)
	}
	ctx := tenant.WithKey(t.Context(), key)
	productionCtx := tenant.WithKey(t.Context(), tenant.Production())
	client := ketoclient.NewClient(fixture.ReadURL(), fixture.WriteURL())
	service := NewKetoService(client)
	manager := NewKetoManager(client)

	for _, test := range []struct {
		name    string
		ctx     context.Context
		subject string
		want    TargetRoles
	}{
		{"inherited administrator", ctx, "production-admin", TargetRoles{Admin: true}},
		{"inherited ban", ctx, "production-banned", TargetRoles{Banned: true}},
		{"local administrator", ctx, "branch-admin", TargetRoles{Admin: true}},
		{"local administrator on production", productionCtx, "branch-admin", TargetRoles{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			roles, err := service.RolesForSubject(test.ctx, test.subject)
			if err != nil {
				t.Fatal(err)
			}

			if roles != test.want {
				t.Errorf("roles=%+v, want %+v", roles, test.want)
			}
		})
	}

	roles, err := service.RolesForSubjects(ctx, []string{"production-admin", "branch-admin", "production-banned"})
	if err != nil {
		t.Fatal(err)
	}
	if roles["production-admin"] != (TargetRoles{}) ||
		roles["production-banned"] != (TargetRoles{}) ||
		roles["branch-admin"] != (TargetRoles{Admin: true}) {
		t.Errorf("branch list should contain only local grants: %v", roles)
	}

	if err := manager.SetBanned(ctx, "tester", true); err != nil {
		t.Fatal(err)
	}
	branchRoles, err := service.RolesForSubject(ctx, "tester")
	if err != nil || !branchRoles.Banned {
		t.Fatalf("local ban=%+v error=%v", branchRoles, err)
	}
	productionRoles, err := service.RolesForSubject(productionCtx, "tester")
	if err != nil || productionRoles.Banned {
		t.Fatalf("local ban changed production: roles=%+v error=%v", productionRoles, err)
	}

	if err := manager.SetBanned(ctx, "tester", false); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetBanned(ctx, "production-banned", false); err != nil {
		t.Fatal(err)
	}
	branchRoles, err = service.RolesForSubject(ctx, "production-banned")
	if err != nil || !branchRoles.Banned {
		t.Fatalf("local unban removed an inherited ban: roles=%+v error=%v", branchRoles, err)
	}
}

func TestRoleLookupsRequireTenant(t *testing.T) {
	fixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := fixture.Reset(t.Context(), "testdata/tenant_relationships.json"); err != nil {
		t.Fatal(err)
	}

	client := ketoclient.NewClient(fixture.ReadURL(), fixture.WriteURL())
	checker := NewKetoChecker(client)
	service := NewKetoService(client)
	manager := NewKetoManager(client)
	for _, ctx := range []context.Context{t.Context(), tenant.WithKey(t.Context(), tenant.Key{})} {
		if allowed, err := checker.CheckAdmin(ctx, "production-admin"); allowed || err == nil {
			t.Errorf("unscoped administrator lookup=(%t, %v)", allowed, err)
		}
		if allowed, err := checker.CheckBanned(ctx, "production-banned"); allowed || err == nil {
			t.Errorf("unscoped ban lookup=(%t, %v)", allowed, err)
		}
		if _, err := service.RolesForSubject(ctx, "production-admin"); err == nil {
			t.Error("unscoped target lookup succeeded")
		}
		if _, err := service.RolesForSubjects(ctx, []string{"production-admin"}); err == nil {
			t.Error("unscoped target list succeeded")
		}
		if err := manager.SetBanned(ctx, "tester", true); err == nil {
			t.Error("unscoped ban write succeeded")
		}
	}
}
