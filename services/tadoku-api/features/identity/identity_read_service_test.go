package identity

import (
	"errors"
	"testing"

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
