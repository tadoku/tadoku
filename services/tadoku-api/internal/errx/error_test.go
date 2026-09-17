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
		name string
		err  *errx.Error
		want string
	}{
		{name: "message", err: &errx.Error{Kind: errx.InvalidInput, Message: "invalid input"}, want: "invalid input"},
		{name: "cause", err: &errx.Error{Kind: errx.Unavailable, Cause: cause}, want: "provider failed"},
		{name: "message and cause", err: &errx.Error{Kind: errx.Unavailable, Message: "permission check", Cause: cause}, want: "permission check: provider failed"},
		{name: "zero value", err: &errx.Error{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.err.Error(); got != test.want {
				t.Errorf("Error()=%q, want %q", got, test.want)
			}
			wrapped := fmt.Errorf("operation: %w", test.err)
			if got := errx.KindOf(wrapped); got != test.err.Kind {
				t.Errorf("KindOf=%v, want %v", got, test.err.Kind)
			}
			if !errors.Is(wrapped, test.err) {
				t.Error("wrapping lost the application error identity")
			}
			if got := test.err.Unwrap(); got != test.err.Cause {
				t.Errorf("Unwrap()=%v, want %v", got, test.err.Cause)
			}
			if test.err.Cause != nil && !errors.Is(wrapped, test.err.Cause) {
				t.Error("wrapping lost the original cause")
			}
		})
	}
}
