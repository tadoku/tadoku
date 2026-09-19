package authz

import (
	"context"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestCurrentUserRoleTreatsEmptySubjectAsGuest(t *testing.T) {
	service := NewService(nil)
	ctx := identity.WithUser(context.Background(), &identity.User{})

	role, err := service.CurrentUserRole(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if role != RoleGuest {
		t.Errorf("role = %q, want %q", role, RoleGuest)
	}
}
