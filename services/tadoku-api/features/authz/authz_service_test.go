package authz

import (
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestCurrentUserRoleTreatsEmptySubjectAsGuest(t *testing.T) {
	service := NewService(nil)
	ctx := identity.WithUser(t.Context(), &identity.User{})

	role, err := service.CurrentUserRole(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if role != RoleGuest {
		t.Errorf("role = %q, want %q", role, RoleGuest)
	}
}

func TestCurrentUserRoleBannedTakesPrecedence(t *testing.T) {
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})
	ctx = permissions.WithBanned(ctx)
	service := NewService(permissions.NewKetoChecker(nil))

	role, err := service.CurrentUserRole(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if role != RoleBanned {
		t.Errorf("role = %q, want %q", role, RoleBanned)
	}
}

func TestPermissionCheckParametersValidate(t *testing.T) {
	tests := []struct {
		name       string
		parameters PermissionCheckParameters
		want       string
	}{
		{
			name:       "namespace is required",
			parameters: PermissionCheckParameters{Object: "tadoku", Relation: "admins"},
			want:       "namespace is required",
		},
		{
			name:       "object is required",
			parameters: PermissionCheckParameters{Namespace: "app", Relation: "admins"},
			want:       "object is required",
		},
		{
			name:       "relation is required",
			parameters: PermissionCheckParameters{Namespace: "app", Object: "tadoku"},
			want:       "relation is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.parameters.Validate()
			if errx.KindOf(err) != errx.InvalidInput {
				t.Errorf("Validate() error kind = %v, want invalid input", errx.KindOf(err))
			}
			if err == nil || err.Error() != test.want {
				t.Errorf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
