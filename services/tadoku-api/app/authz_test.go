package app

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestProxyAdminCheckRequiresCallbackAuthenticationBeforeValidation(t *testing.T) {
	application := New(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := application.ProxyAdminCheck(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.Unauthorized {
		t.Errorf("ProxyAdminCheck() error kind = %v, want unauthorized", errx.KindOf(err))
	}
}
