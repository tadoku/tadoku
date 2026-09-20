package authz

import (
	"context"
	"errors"
	"testing"

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
	adminLookups := 0
	service := NewService(permissions.NewChecker(func(context.Context, string) (bool, error) {
		adminLookups++
		return true, nil
	}))

	role, err := service.CurrentUserRole(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if role != RoleBanned {
		t.Errorf("role = %q, want %q", role, RoleBanned)
	}
	if adminLookups != 0 {
		t.Errorf("admin lookups = %d, want 0", adminLookups)
	}
}

func TestPermissionCheckParametersValidate(t *testing.T) {
	tests := []struct {
		name       string
		parameters PermissionCheckParameters
		want       error
	}{
		{
			name:       "namespace is required",
			parameters: PermissionCheckParameters{Object: "tadoku", Relation: "admins"},
			want:       ErrPermissionNamespaceRequired,
		},
		{
			name:       "object is required",
			parameters: PermissionCheckParameters{Namespace: "app", Relation: "admins"},
			want:       ErrPermissionObjectRequired,
		},
		{
			name:       "relation is required",
			parameters: PermissionCheckParameters{Namespace: "app", Object: "tadoku"},
			want:       ErrPermissionRelationRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.parameters.Validate(); !errors.Is(err, test.want) {
				t.Errorf("Validate() error = %v, want %v", err, test.want)
			}
		})
	}
}
