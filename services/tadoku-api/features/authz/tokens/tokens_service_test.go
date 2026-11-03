package tokens

import (
	"errors"
	"testing"

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
