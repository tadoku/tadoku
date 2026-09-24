package authz

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestCurrentUserRoleTreatsEmptySubjectAsGuest(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil)
	ctx := identity.WithUser(t.Context(), &identity.User{})

	role, err := service.CurrentUserRole(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if role != RoleGuest {
		t.Errorf("role = %q, want %q", role, RoleGuest)
	}
}

func TestProxyAdminCheckRejectsNilSubject(t *testing.T) {
	service := NewService(nil, nil, nil, nil, nil)

	_, err := service.ProxyAdminCheck(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.InvalidInput {
		t.Errorf("ProxyAdminCheck() error kind = %v, want invalid input", errx.KindOf(err))
	}
	if err == nil || err.Error() != "subject must be a UUID" {
		t.Errorf("ProxyAdminCheck() error = %v, want %q", err, "subject must be a UUID")
	}
}

func TestCurrentUserRoleBannedTakesPrecedence(t *testing.T) {
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})
	ctx = permissions.WithBanned(ctx)
	service := NewService(permissions.NewKetoChecker(nil), nil, nil, nil, nil)

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

func TestRoleUpdateParametersValidation(t *testing.T) {
	t.Parallel()

	valid := RoleUpdateParameters{
		UserID: uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Role:   RoleBanned,
		Reason: "moderation reason",
	}
	for _, test := range []struct {
		name   string
		change func(*RoleUpdateParameters)
		want   string
	}{
		{
			name: "unsupported role",
			change: func(parameters *RoleUpdateParameters) {
				parameters.Role = RoleAdmin
			},
			want: "role must be 'user' or 'banned'",
		},
		{
			name: "empty reason",
			change: func(parameters *RoleUpdateParameters) {
				parameters.Reason = ""
			},
			want: "reason is required",
		},
		{
			name: "reason over byte limit",
			change: func(parameters *RoleUpdateParameters) {
				parameters.Reason = strings.Repeat("é", 501)
			},
			want: "reason must be at most 1000 bytes",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)

			err := parameters.Validate()
			if errx.KindOf(err) != errx.InvalidInput {
				t.Errorf("Validate() error kind = %v, want invalid input", errx.KindOf(err))
			}
			if err == nil || err.Error() != test.want {
				t.Errorf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}

	for _, reason := range []string{
		" ",
		strings.Repeat("a", 1000),
		strings.Repeat("é", 500),
	} {
		parameters := valid
		parameters.Reason = reason
		if err := parameters.Validate(); err != nil {
			t.Errorf("Validate() rejected %d-byte reason: %v", len(reason), err)
		}
	}

	valid.Role = RoleUser
	if err := valid.Validate(); err != nil {
		t.Errorf("Validate() rejected user role: %v", err)
	}
}
