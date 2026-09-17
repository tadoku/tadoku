package errx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestError(t *testing.T) {
	t.Parallel()
	cause := errors.New("provider failed")
	for _, test := range []struct {
		name      string
		err       *errx.Error
		want      string
		wantKind  errx.Kind
		wantCause error
	}{
		{name: "invalid input", err: errx.NewInvalidInputError("invalid input"), want: "invalid input", wantKind: errx.InvalidInput},
		{name: "unauthorized", err: errx.NewUnauthorizedError("unauthorized"), want: "unauthorized", wantKind: errx.Unauthorized},
		{name: "forbidden", err: errx.NewForbiddenError("forbidden"), want: "forbidden", wantKind: errx.Forbidden},
		{name: "not found", err: errx.NewNotFoundError("not found"), want: "not found", wantKind: errx.NotFound},
		{name: "unavailable without cause", err: errx.NewUnavailableError("unavailable", nil), want: "unavailable", wantKind: errx.Unavailable},
		{name: "cause", err: errx.NewUnavailableError("", cause), want: "provider failed", wantKind: errx.Unavailable, wantCause: cause},
		{name: "message and cause", err: errx.NewUnavailableError("permission check", cause), want: "permission check: provider failed", wantKind: errx.Unavailable, wantCause: cause},
		{name: "zero value", err: &errx.Error{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.err.Error(); got != test.want {
				t.Errorf("Error()=%q, want %q", got, test.want)
			}
			wrapped := fmt.Errorf("operation: %w", test.err)
			if got := errx.KindOf(wrapped); got != test.wantKind {
				t.Errorf("KindOf=%v, want %v", got, test.wantKind)
			}
			if !errors.Is(wrapped, test.err) {
				t.Error("wrapping lost the application error identity")
			}
			if got := test.err.Unwrap(); got != test.wantCause {
				t.Errorf("Unwrap()=%v, want %v", got, test.wantCause)
			}
			if test.wantCause != nil && !errors.Is(wrapped, test.wantCause) {
				t.Error("wrapping lost the original cause")
			}
		})
	}
}
