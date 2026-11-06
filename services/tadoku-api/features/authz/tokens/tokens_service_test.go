package tokens

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreateRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, false).Create(t.Context(), Create{})
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("Create without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestListRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, false).List(t.Context())
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("List without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestRevokeRequiresTenant(t *testing.T) {
	t.Parallel()

	err := NewService(nil, false).Revoke(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("Revoke without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}
