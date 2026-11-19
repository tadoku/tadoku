package identity

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestListAccountsRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewReader(nil).ListAccounts(t.Context(), ListAccounts{})
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("ListAccounts without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestFindAccountRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewReader(nil).FindAccount(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("FindAccount without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestUserExistsRequiresTenant(t *testing.T) {
	t.Parallel()

	exists, err := NewReader(nil).UserExists(t.Context(), uuid.Nil)
	if exists || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("UserExists without tenant = (%v, %v), want (false, unavailable wrapping %v)", exists, err, postgres.ErrNoTenant)
	}
}
