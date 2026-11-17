package api

import (
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestAuthenticateEmailReaderUnavailableWithoutTokens(t *testing.T) {
	t.Parallel()

	_, err := New(Dependencies{}).AuthenticateEmailReader(t.Context(), "anything")
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("AuthenticateEmailReader without tokens = %v, want unavailable", err)
	}
}
